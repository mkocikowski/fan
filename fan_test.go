package main

import (
	//"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"
)

var _ = fmt.Println

func TestSpawn(t *testing.T) {
	s := "foo"
	stdin, stdout := spawnWorker([]string{"cat"})
	stdin.Write([]byte(s))
	stdin.Close()
	b := make([]byte, len(s))
	_, err := stdout.Read(b)
	if err != nil {
		t.Error(err)
	}
	if string(b) != s {
		t.Errorf("expected '%s' got '%s'", s, string(b))
	}
}

func TestAttach(t *testing.T) {
	s := "foo\n"
	wOut, wIn := io.Pipe()
	inChan := make(chan []byte)
	outChan := make(chan []byte)
	go attachWorker(wIn, wOut, inChan, outChan)
	inChan <- []byte(s)
	close(inChan) // flush the buffer, otherwise deadlock
	b := <-outChan
	if string(b) != s {
		t.Errorf("expected '%s' got '%s'", s, string(b))
	}
}

func TestPool(t *testing.T) {
	s := "foo\n"
	inChan, outChan := createWorkerPool(1, []string{"cat"})
	inChan <- []byte(s)
	close(inChan) // flush the buffer, otherwise deadlock
	b := <-outChan
	if string(b) != s {
		t.Errorf("expected '%s' got '%s'", s, string(b))
	}
}

func TestRun(t *testing.T) {
	s := "foo\n"
	in := bytes.NewBufferString(s)
	var out bytes.Buffer
	run(1, []string{"cat"}, in, &out)
	b := out.Bytes()
	if string(b) != s {
		t.Errorf("expected '%s' got '%s'", s, string(b))
	}
}
