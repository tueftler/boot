package main

import (
	"bytes"
	"fmt"
)

var Print = func(arg string) { fmt.Print(arg) }

type Stream struct {
	prefix  string
	write   func(string)
	started bool
}

func NewStream(prefix string, write func(string)) *Stream {
	return &Stream{prefix: prefix, write: write, started: false}
}

func (s *Stream) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Trim off byte sequences [1 0 0 0 0 0 0 {1,9,10}]
	if p[0] == '\001' {
		s.Write(p[8:len(p)])
		return len(p), nil
	}

	if !s.started {
		s.write(s.prefix)
		// DEBUG s.write(fmt.Sprintf("%+v\n", p))
		// DEBUG s.write(s.prefix)
		s.started = true
	}

	pos := bytes.IndexByte(p, '\n')
	if pos == -1 {
		s.write(string(p))
	} else {
		pos++
		s.write(string(p[0:pos]))
		s.started = false
		s.Write(p[pos:len(p)])
	}

	return len(p), nil
}
