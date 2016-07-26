package main

import (
	"bytes"
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestStartWorker(t *testing.T) {
	var tests = []struct {
		cmd      string
		input    string
		expected string
		err      string
	}{
		{"cat", "foo\n", "foo\n", ""},
		{"cat", "foo\nbar", "foo\n", ""}, // missing second newline, will not get sent
		{"foo", "foo\n", "", "*exec.Error"},
		{"false", "foo\n", "", "*exec.ExitError"},
	}

	for _, test := range tests {
		ic := make(chan []byte)
		oc := make(chan []byte)
		failChan := make(chan error)
		// check for errors starting a worker
		err := startWorker([]string{test.cmd}, ic, oc, failChan)
		if err != nil {
			if s := fmt.Sprintf("%T", err); s != test.err {
				t.Fatalf("expected '%q' got '%q'", test.err, s)
			}
			continue
		}
		ic <- []byte(test.input)
		close(ic) // flushes input
		select {
		case b := <-oc:
			if string(b) != test.expected {
				t.Errorf("expected '%q' got '%q'", test.expected, string(b))
			}
		case e := <-failChan:
			if s := fmt.Sprintf("%T", e); s != test.err {
				t.Fatalf("expected '%q' got '%q'", test.err, s)
			}
		}
	}
}

func TestRun(t *testing.T) {
	s := "foo\nbar\nbaz"
	in := bytes.NewBufferString(s)
	var out bytes.Buffer
	err := run(1, []string{"cat"}, in, &out)
	if err != nil {
		t.Error("got an error where expected none", err)
	}
	b := out.Bytes()
	expect := "foo\nbar\n"
	if string(b) != expect {
		t.Errorf("expected '%q' got '%q'", expect, string(b))
	}
}
