package main

import (
	"bufio"
	"os/exec"
)

type Worker struct {
	Args    []string
	Input   chan []byte
	Output  chan []byte
	Fail    chan error
	running bool
}

func (w *Worker) Start() error {

	cmd := exec.Command(w.Args[0], w.Args[1:]...)
	stdi, _ := cmd.StdinPipe()
	stdo, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	// collect worker's output; exit when unable to read from worker's stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerOutputBuf := bufio.NewReader(stdo)
		for {
			b, err := workerOutputBuf.ReadBytes('\n')
			if err != nil {
				w.Fail <- cmd.Wait()
				return
			}
			w.Output <- b
		}
	}()

	// feed data to worker's input; exit if unable to write, or when the input channel closed
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range w.Input {
			_, err := stdi.Write(line)
			if err != nil {
				w.Fail <- cmd.Wait()
				return
			}
		}
		// closing proces' stdin will make it exit, if it is reading from stdin
		// so this is a way to gracefuly close a worker process
		stdi.Close()
	}()

	return nil
}
