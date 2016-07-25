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
	wg        = new(sync.WaitGroup)
)

func startWorker(args []string, iChan, oChan chan []byte) {
	cmd := exec.Command(args[0], args[1:]...)
	stdi, _ := cmd.StdinPipe()
	stdo, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		workerOutputBuf := bufio.NewReader(stdo)
		for {
			b, err := workerOutputBuf.ReadBytes('\n')
			if err != nil { // most likely EOF
				if err := cmd.Wait(); err != nil {
					log.Fatal(err)
				}
				break
			}
			oChan <- b
		}
	}()
	go func() {
		defer wg.Done()
		workerInputBuf := bufio.NewWriter(stdi)
		for line := range iChan {
			_, err := workerInputBuf.Write(line)
			if err != nil {
				log.Fatal(err)
			}
		}
		workerInputBuf.Flush()
		stdi.Close() // closing proces' stdin will make it exit
	}()
}

func run(n int, args []string, r io.Reader, w io.Writer) {
	iChan := make(chan []byte)
	oChan := make(chan []byte)
	for i := 0; i < n; i++ {
		startWorker(args, iChan, oChan)
	}
	go func() {
		inputBuf := bufio.NewReader(r)
		for {
			b, err := inputBuf.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			iChan <- b
		}
		close(iChan)
	}()
	go func() {
		wg.Wait()
		// all the workers have exited, so there is no one writing to the
		// oChan, so close it to signal completion and have run() break
		close(oChan)
	}()
	outputBuf := bufio.NewWriter(w)
	defer outputBuf.Flush()
	for line := range oChan {
		outputBuf.Write(line)
	}
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
