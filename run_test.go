package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestFanOut(t *testing.T) {
	iChan := make(chan []byte)
	errChan := make(chan error)
	r := bytes.NewBufferString("foo\nbar\n")
	go fanOut(r, iChan, errChan)
	if b := <-iChan; string(b) != "foo\n" {
		t.Fatalf("expected %q got %q", "foo\n", string(b))
	}
	errChan <- errors.New("foo error")
	if _, ok := <-iChan; ok {
		t.Fatal("expected channel to be closed")
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
