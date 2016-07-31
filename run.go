package main

import (
	"bufio"
	"io"
	"log"
	"sync"
)

func fanOut(r io.Reader, iChan chan<- []byte, errChan <-chan error) {
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
}

func run(n int, args []string, r io.Reader, w io.Writer) error {

	iChan := make(chan []byte)
	oChan := make(chan []byte)
	errChan := make(chan error, n*2)
	wg := new(sync.WaitGroup)

	for i := 0; i < n; i++ {
		if err := startWorker(args, wg, iChan, oChan, errChan); err != nil {
			return err
		}
	}

	// send data from fan's stdin to iChan (to workers)
	go fanOut(r, iChan, errChan)

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
