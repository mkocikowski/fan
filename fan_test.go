package main

import (
	//"bufio"
	"bytes"
	"fmt"
	//	"io"
	//	"sync"
	"testing"
)

var _ = fmt.Println

//func TestExecWorker(t *testing.T) {
//	s := "foo"
//	stdin, stdout := execWorker([]string{"cat"})
//	stdin.Write([]byte(s))
//	stdin.Close()
//	b := make([]byte, len(s))
//	_, err := stdout.Read(b)
//	if err != nil {
//		t.Error(err)
//	}
//	if string(b) != s {
//		t.Errorf("expected '%q' got '%q'", s, string(b))
//	}
//}

func TestRun(t *testing.T) {
	s := "foo\nbar\nbaz"
	in := bytes.NewBufferString(s)
	var out bytes.Buffer
	run(1, []string{"cat"}, newProcessWorker, in, &out)
	b := out.Bytes()
	expect := "foo\nbar\n"
	if string(b) != expect {
		t.Errorf("expected '%q' got '%q'", expect, string(b))
	}
}
