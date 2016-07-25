package main

import (
	"bytes"
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestStartWorker(t *testing.T) {
	cmd := []string{"cat"}
	ic := make(chan []byte)
	oc := make(chan []byte)
	startWorker(cmd, ic, oc)
	s := "foo\n"
	ic <- []byte(s)
	close(ic) // flushes input
	b := <-oc
	if string(b) != s {
		t.Errorf("expected '%q' got '%q'", s, string(b))
	}
}

func TestRun(t *testing.T) {
	s := "foo\nbar\nbaz"
	in := bytes.NewBufferString(s)
	var out bytes.Buffer
	run(1, []string{"cat"}, in, &out)
	b := out.Bytes()
	expect := "foo\nbar\n"
	if string(b) != expect {
		t.Errorf("expected '%q' got '%q'", expect, string(b))
	}
}
