package output

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

// NewStream creates a stream with a given prefix and writer
func NewStream(prefix string, write func(string)) *Stream {
	return &Stream{prefix: prefix, write: write, started: false}
}

// NewStream creates a stream on the same writer as this stream, but
// with a different prefix
func (s *Stream) Prefixed(prefix string) *Stream {
	return &Stream{prefix: prefix, write: s.write, started: false}
}

// Printf formats arguments without any coloring
func (s *Stream) Printf(format string, args ...interface{}) {
	fmt.Fprintf(s, format, args...)
}

// Println prints a line without any coloring
func (s *Stream) Println(args ...interface{}) {
	fmt.Fprintln(s, args...)
}

// Line formats arguments and prints it
func (s *Stream) Line(kind, format string, args ...interface{}) {
	fmt.Fprintf(s, Line(kind, format), args...)
}

// Write writes the given bytes, prefixing all lines with the given prefix
func (s *Stream) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if !s.started {
		s.write(s.prefix)
		s.started = true
	}

	pos := bytes.IndexByte(p, '\n')
	if pos == -1 {
		s.write(string(p))
	} else {
		pos++
		s.write(string(p[0:pos]))
		s.started = false
		s.Write(p[pos:])
	}

	return len(p), nil
}
