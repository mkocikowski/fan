package main

import (
	"bufio"
	"flag"
	"fmt"
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
)

func worker(in, out chan []byte, args []string, wg *sync.WaitGroup) {

	cmd := exec.Command(args[0], args[1:]...)
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
			out <- c
		}
		// by now the enclosing goroutine would have exited having exhausted the input
		cmd.Wait()
	}()

	for line := range in {
		i, err := cmdStdin.Write(line)
		if err != nil {
			log.Fatal(i, err)
		}
	}
	// closing command's stdin will make it exit once done processing
	cmdStdin.Close()
	// inner goroutine will continue reading command stdout until exhausted
}

func run(n int, args []string) {

	if n < 1 {
		log.Fatal("n must be > 0")
	}
	if len(args) < 1 {
		log.Fatal("command missing")
	}

	var wg sync.WaitGroup
	in := make(chan []byte)
	out := make(chan []byte)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(in, out, args, &wg)
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
			in <- c
		}
		close(in)  // signal to workers to wrap things up
		wg.Wait()  // wait for all workers
		close(out) // signal completion to exit main loop
	}()

	// this when exit when out chan is closed by the goroutine above
	for line := range out {
		os.Stdout.Write(line)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "v%s %s %s\n", Version, BuildHash, BuildDate)
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
	run(n, flag.Args())
}
