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
	Version   = "0.2.0"
	wg        = new(sync.WaitGroup)
)

func startWorker(args []string, iChan, oChan chan []byte, failChan chan<- error) error {

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
			if err != nil { // most likely EOF
				if err := cmd.Wait(); err != nil {
					failChan <- err
				}
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
				failChan <- err
				return
			}
		}
		stdi.Close() // closing proces' stdin will make it exit, if it is reading from stdin
	}()

	return nil
}

func run(n int, args []string, r io.Reader, w io.Writer) error {

	iChan := make(chan []byte)
	oChan := make(chan []byte)
	failChan := make(chan error)

	for i := 0; i < n; i++ {
		if err := startWorker(args, iChan, oChan, failChan); err != nil {
			return err
		}
	}

	// send data from fan's stdin to iChan
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
			// whichever comes first: a worker reading from the unbuffered channel, or an error coming from a worker
			select {
			case iChan <- b:
			case err := <-failChan:
				log.Println(err)
				break
			}
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
