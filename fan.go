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

func spawnWorker(args []string) (io.WriteCloser, io.ReadCloser) {
	cmd := exec.Command(args[0], args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	return stdin, stdout
}

func attachWorker(workerInput io.WriteCloser, workerOutput io.Reader, dataIn, dataOut chan []byte) {
	go func() {
		defer wg.Done()
		workerOutputBuf := bufio.NewReader(workerOutput)
		for {
			b, err := workerOutputBuf.ReadBytes('\n')
			if err != nil {
				break // TODO: deal with partial reads, errors?
			}
			c := make([]byte, len(b))
			copy(c, b)
			dataOut <- c
		}
	}()
	workerInputBuf := bufio.NewWriter(workerInput)
	for line := range dataIn {
		_, err := workerInputBuf.Write(line)
		if err != nil {
			log.Fatal(err)
		}
	}
	workerInputBuf.Flush()
	workerInput.Close() // closing command's stdin will make it exit once done processing
}

func createWorkerPool(n int, args []string) (chan []byte, chan []byte) {
	inChan := make(chan []byte)
	outChan := make(chan []byte)
	for i := 0; i < n; i++ {
		wIn, wOut := spawnWorker(args)
		go attachWorker(wIn, wOut, inChan, outChan)
	}
	return inChan, outChan
}

func feedDataToWorkerPool(fanInput io.Reader, workersInputChan chan []byte) {
	inputBuf := bufio.NewReader(fanInput)
	for {
		b, err := inputBuf.ReadBytes('\n')
		if err != nil {
			break
		}
		c := make([]byte, len(b))
		copy(c, b)
		workersInputChan <- c
	}
	close(workersInputChan)
}

func run(n int, args []string, fanInput io.Reader, fanOutput io.Writer) {
	workersInputChan, workersOutputChan := createWorkerPool(n, args)
	go feedDataToWorkerPool(fanInput, workersInputChan)
	go func() {
		wg.Wait()
		// all the workers have exited, so there is no one writing to the
		// output chan, so close it to signal completion
		close(workersOutputChan)
	}()
	outputBuf := bufio.NewWriter(fanOutput)
	defer outputBuf.Flush()
	// this when exit when output chan is closed by the goroutine above
	for line := range workersOutputChan {
		outputBuf.Write(line)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "v%s %s %s %s\n", Version, BuildHash, BuildDate, runtime.Version())
		fmt.Fprintf(os.Stderr, "usage: fan [-n numprocs] command\n")
		flag.PrintDefaults()
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
