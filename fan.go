package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

var (
	BuildHash string
	BuildDate string
	Version   = "0.1.0"
	wg        sync.WaitGroup
)

type worker interface {
	work(wg *sync.WaitGroup)
	getErr() error
}

type processWorker struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	input  chan []byte
	output chan []byte
	cmd    *exec.Cmd
	err    error
}

type workerFactory func(args []string, input, output chan []byte) (*processWorker, error)

func newProcessWorker(args []string, input, output chan []byte) (*processWorker, error) {
	cmd := exec.Command(args[0], args[1:]...)
	w := processWorker{
		cmd:    cmd,
		input:  input,
		output: output,
	}
	w.stdin, _ = cmd.StdinPipe()
	w.stdout, _ = cmd.StdoutPipe()
	err := w.cmd.Start()
	return &w, err
}

func (w *processWorker) work(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		go func() {
			defer wg.Done()
			workerOutputBuf := bufio.NewReader(w.stdout)
			for {
				b, err := workerOutputBuf.ReadBytes('\n')
				if err != nil { // most likely EOF
					if err := w.cmd.Wait(); err != nil {
						log.Fatal(err)
					}
					break
				}
				w.output <- b
			}
		}()
		workerInputBuf := bufio.NewWriter(w.stdin)
		for line := range w.input {
			_, err := workerInputBuf.Write(line)
			if err != nil {
				log.Fatal(err)
			}
		}
		workerInputBuf.Flush()
		w.stdin.Close() // closing command's stdin will make it exit once done processing
	}()
}

func (w *processWorker) getErr() error {
	return w.err
}

type workerPool struct {
	input  chan []byte
	output chan []byte
	errors chan []byte
	n      int
	args   []string
	wg     *sync.WaitGroup
}

func newWorkerPool(n int, args []string, f workerFactory) *workerPool {
	pool := &workerPool{
		input:  make(chan []byte),
		output: make(chan []byte),
		n:      n,
		args:   args,
		wg:     new(sync.WaitGroup),
	}
	for i := 0; i < n; i++ {
		worker, err := f(args, pool.input, pool.output)
		if err != nil {
			log.Fatal(err)
		}
		worker.work(pool.wg)
	}
	return pool
}

func (pool *workerPool) work(fanInput io.Reader) {
	go func() {
		inputBuf := bufio.NewReader(fanInput)
		for {
			// error when fanInput closed / exhausted
			b, err := inputBuf.ReadBytes('\n')
			if err != nil {
				break
			}
			pool.input <- b
		}
		close(pool.input)
	}()
}

func (pool *workerPool) wait(fanOutput io.Writer) {
	go func() {
		pool.wg.Wait()
		// all the workers have exited, so there is no one writing to the
		// output chan, so close it to signal completion
		close(pool.output)
	}()
	outputBuf := bufio.NewWriter(fanOutput)
	defer outputBuf.Flush()
	// this when exit when output chan is closed by the goroutine above
	for line := range pool.output {
		outputBuf.Write(line)
	}
}

func run(n int, args []string, f workerFactory, input io.Reader, output io.Writer) {
	pool := newWorkerPool(n, args, f)
	pool.work(input)
	pool.wait(output)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "v%s %s %s %s\n", Version, BuildHash, BuildDate, runtime.Version())
		fmt.Fprintf(os.Stderr, "usage: fan [-n numprocs] command\n")
		flag.PrintDefaults()
	}
	n := flag.Int("n", runtime.NumCPU(), "number of processes to run")
	flag.Parse()
	if *n < 1 {
		fmt.Fprintln(os.Stderr, "error: n must be > 0")
		os.Exit(1)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	run(*n, flag.Args(), newProcessWorker, os.Stdin, os.Stdout)
}
