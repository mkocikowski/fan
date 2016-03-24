package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

var (
	inChan  = make(chan []byte)
	outChan = make(chan []byte)
	wg      sync.WaitGroup
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: np [-n numprocs] command\n")
		flag.PrintDefaults()
	}
}

func worker() {
	defer wg.Done()

	cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()

	output := bufio.NewScanner(cmdStdout)
	go func() {
		for output.Scan() {
			b := output.Bytes()
			c := make([]byte, len(b), len(b)+1)
			copy(c, b)
			outChan <- append(c, '\n')
		}
	}()

	for line := range inChan {
		i, err := cmdStdin.Write(line)
		if err != nil {
			log.Println(i, err)
		}
	}

	cmdStdin.Close() // signal for child process to exit
	cmd.Wait()
}

func main() {

	var n int
	flag.IntVar(&n, "n", 1, "number of processes to run")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			b := input.Bytes()
			c := make([]byte, len(b), len(b)+1)
			copy(c, b)
			inChan <- append(c, '\n')
		}
		close(inChan)  // signal to workers to exit
		wg.Wait()      // wait for all workers
		close(outChan) // signal completion to exit main loop
	}()

	// this when exit when outChan closed by the goroutine above
	for line := range outChan {
		os.Stdout.Write(line)
	}
}
