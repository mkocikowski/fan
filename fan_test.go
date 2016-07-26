package main

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

var _ = fmt.Println

func TestSpawnWorker(t *testing.T) {
	var tests = []struct {
		cmd string
		err bool
	}{
		{"cat", false},
		{"foo", true},
	}
	for _, test := range tests {
		err := startWorker([]string{test.cmd}, nil, nil, nil)
		if err == nil == test.err {
			t.Error("expected error")
		}
	}
}

func TestStartWorker(t *testing.T) {
	var tests = []struct {
		cmd      string
		input    string
		expected string
		err      string
	}{
		{"cat", "foo\n", "foo\n", ""},
		{"cat", "foo\nbar", "foo\n", ""},
		{"false", "foo\n", "", "*exec.ExitError"},
	}
	for _, test := range tests {
		wg = new(sync.WaitGroup)
		ic := make(chan []byte)
		oc := make(chan []byte)
		errChan := make(chan error, 2)
		err := startWorker([]string{test.cmd}, ic, oc, errChan)
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		ic <- []byte(test.input)
		close(ic) // flushes input
		select {
		case b := <-oc:
			if string(b) != test.expected {
				t.Errorf("expected '%q' got '%q'", test.expected, string(b))
			}
		case e := <-errChan:
			if s := fmt.Sprintf("%T", e); s != test.err {
				t.Fatalf("expected '%q' got '%q'", test.err, s)
			}
		}
	}
}

func TestRun(t *testing.T) {
	wg = new(sync.WaitGroup)
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
