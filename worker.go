package main

import (
	"bufio"
	"os/exec"
	"sync"
)

func startWorker(args []string, wg *sync.WaitGroup, iChan, oChan chan []byte, errChan chan<- error) error {

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
			if err != nil {
				errChan <- cmd.Wait()
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
				errChan <- cmd.Wait()
				return
			}
		}
		// closing proces' stdin will make it exit, if it is reading from stdin
		// so this is a way to gracefuly close a worker process
		stdi.Close()
	}()

	return nil
}
