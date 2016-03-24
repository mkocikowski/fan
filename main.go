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

func run(in, out chan []byte, wg *sync.WaitGroup) {

	wg.Add(1)
	defer wg.Done()

	cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	go func() {
		wg.Add(1)
		defer wg.Done()
		for scanner.Scan() {
			b := scanner.Bytes()
			c := make([]byte, len(b), len(b)+1)
			copy(c, b)
			out <- append(c, '\n')
		}
	}()

	for b := range in {
		i, err := stdin.Write(b)
		if err != nil {
			log.Println(i, err)
		}
	}

	stdin.Close() // signal for child process to exit
	cmd.Wait()
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: np [-n numprocs] command\n")
		flag.PrintDefaults()
	}
}

func main() {

	var n int
	flag.IntVar(&n, "n", 1, "number of processes to run")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	in := make(chan []byte)
	out := make(chan []byte)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		go run(in, out, &wg)
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			b := scanner.Bytes()
			c := make([]byte, len(b), len(b)+1)
			copy(c, b)
			in <- append(c, '\n')
		}
		close(in)
		wg.Wait() // wait for all children to finish
		close(out)
	}()

	for b := range out {
		os.Stdout.Write(b)
	}
}
