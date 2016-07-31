package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	BuildHash string
	BuildDate string
	Version   = "0.2.0"
)

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
