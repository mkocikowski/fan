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
		fmt.Fprintf(os.Stderr, "usage: fan [-n numprocs] command\n")
		flag.PrintDefaults()
	}
}

func worker() {

	cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1) // removed in goroutine below
	output := bufio.NewReader(cmdStdout)
	go func() {
		defer wg.Done()
		// this will keep on going until the command's stdout is exhausted
		for {
			b, err := output.ReadBytes('\n')
			if err != nil {
				break
			}
			c := make([]byte, len(b))
			copy(c, b)
			outChan <- c
		}
		// by now the enclosing goroutine would have exited having exhausted the input
		cmd.Wait()
	}()

	for line := range inChan {
		i, err := cmdStdin.Write(line)
		if err != nil {
			log.Fatal(i, err)
		}
	}
	// closing command's stdin will make it exit once done processing
	cmdStdin.Close()
}

func main() {

	var n int
	flag.IntVar(&n, "n", 1, "number of processes to run")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if n < 1 {
		fmt.Fprintln(os.Stderr, "n must be > 0")
		os.Exit(1)
	}

	for i := 0; i < n; i++ {
		go worker()
	}

	go func() {
		input := bufio.NewReader(os.Stdin)
		for {
			b, err := input.ReadBytes('\n')
			if err != nil {
				break
			}
			c := make([]byte, len(b))
			copy(c, b)
			inChan <- c
		}
		close(inChan)  // signal to workers to wrap things up
		wg.Wait()      // wait for all workers
		close(outChan) // signal completion to exit main loop
	}()

	// this when exit when outChan closed by the goroutine above
	for line := range outChan {
		os.Stdout.Write(line)
	}
}
