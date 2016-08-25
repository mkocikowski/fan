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
	Version   = "0.2.1"
	wg        = new(sync.WaitGroup)
)

func startWorker(args []string, iChan, oChan chan []byte, errChan chan<- error) error {

	cmd := exec.Command(args[0], args[1:]...)
	stdi, _ := cmd.StdinPipe()
	stdo, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	wg.Add(1)
	// collect worker's output; exit when unable to read from worker's stdout
	go func() {
		defer wg.Done()
		workerOutputBuf := bufio.NewReader(stdo)
		for {
			b, err := workerOutputBuf.ReadBytes('\n')
			if err != nil {
				errChan <- cmd.Wait()
				return
			}
			oChan <- b
		}
	}()

	wg.Add(1)
	// feed data to worker's input; exit if unable to write, or when the input channel closed
	go func() {
		defer wg.Done()
		for line := range iChan {
			_, err := stdi.Write(line)
			if err != nil {
				errChan <- cmd.Wait()
				return
			}
		}
		// closing proces' stdin will make it exit, if it is reading from stdin
		// so this is a way to gracefuly close a worker process
		stdi.Close()
	}()

	return nil
}

func run(n int, args []string, r io.Reader, w io.Writer) error {

	iChan := make(chan []byte)
	oChan := make(chan []byte)
	errChan := make(chan error, n*2)

	for i := 0; i < n; i++ {
		if err := startWorker(args, iChan, oChan, errChan); err != nil {
			return err
		}
	}

	// send data from fan's stdin to iChan
	go func() {
		inputBuf := bufio.NewReader(r)
		for {
			b, err := inputBuf.ReadBytes('\n')
			// ReadBytes returns err != nil if and only if the returned data does not end in delim
			if err != nil {
				break
			}
			// whichever comes first: a worker reading from the unbuffered channel, or an error coming from a worker
			select {
			case iChan <- b:
			case err := <-errChan:
				log.Println(err)
				break
			}
		}
		// when workers try to read on closed channel, they will exit
		close(iChan)
	}()

	go func() {
		// wait for all workers to exit
		wg.Wait()
		// closing output channel will terminate run's main loop
		close(oChan)
	}()

	outputBuf := bufio.NewWriter(w)
	defer outputBuf.Flush()
	for line := range oChan {
		outputBuf.Write(line)
	}

	return nil
}

const (
	usage = `
fan splits its input on newline and sends individual lines to worker processes
(specified with 'command') on stdin. It atomically combines stdout from the
worker processes (line by line) and outputs it on stdout. Error in a worker
process results in fan shutting down all workers and exiting.`
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "v%s %s %s %s\n", Version, BuildHash, BuildDate, runtime.Version())
		fmt.Fprintf(os.Stderr, "usage: fan [-n numprocs] command\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, usage)
	}
	var n int
	flag.IntVar(&n, "n", runtime.NumCPU(), "number of processes to run")
	flag.Parse()
	if n < 1 {
		fmt.Fprintln(os.Stderr, "error: n must be > 0")
		os.Exit(1)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	run(n, flag.Args(), os.Stdin, os.Stdout)
}
