package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

func run(in, out chan string, wg *sync.WaitGroup) {

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
			out <- scanner.Text()
		}
		log.Println("exiting scanner")
	}()

	for s := range in {
		i, err := io.WriteString(stdin, s+"\n")
		if err != nil {
			log.Println(i, err)
		}
	}

	stdin.Close() // make child process exit
	cmd.Wait()
	log.Printf("exiting %v\n", cmd.Process)

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

	var wg sync.WaitGroup
	in := make(chan string)
	out := make(chan string)
	for i := 0; i < n; i++ {
		go run(in, out, &wg)
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			t := scanner.Text()
			in <- t
		}
		close(in)
		wg.Wait() // wait for all children to finish
		close(out)
	}()

	for s := range out {
		fmt.Println(s)
	}
	log.Println("out of out")

}
