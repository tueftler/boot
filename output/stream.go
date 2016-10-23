package output

import (
	"bytes"
	"fmt"
)

var Print = func(arg string) { fmt.Print(arg) }

type Stream struct {
	Prefix  string
	writer  func(string)
	started bool
}

// NewStream creates a stream with a given prefix and writer
func NewStream(prefix string, writer func(string)) *Stream {
	return &Stream{Prefix: prefix, writer: writer, started: false}
}

// Prefixed creates a stream on the same writer as this stream, but
// with a different prefix
func (s *Stream) Prefixed(prefix string) *Stream {
	return &Stream{Prefix: prefix, writer: s.writer, started: false}
}

// Printf formats arguments without any coloring
func (s *Stream) Printf(format string, args ...interface{}) {
	fmt.Fprintf(s, format, args...)
}

// Println prints a line without any coloring
func (s *Stream) Println(args ...interface{}) {
	fmt.Fprintln(s, args...)
}

// Line formats arguments and prints result
func (s *Stream) Line(kind, format string, args ...interface{}) {
	fmt.Fprintf(s, Text(kind, format)+"\n", args...)
}

// Info formats arguments as information and prints result
func (s *Stream) Info(format string, args ...interface{}) {
	fmt.Fprintf(s, Text("info", format)+"\n", args...)
}

// Error formats arguments as information and prints result
func (s *Stream) Error(format string, args ...interface{}) {
	fmt.Fprintf(s, Text("error", format)+"\n", args...)
}

// Warning formats arguments as information and prints result
func (s *Stream) Warning(format string, args ...interface{}) {
	fmt.Fprintf(s, Text("warning", format)+"\n", args...)
}

// Success formats arguments as information and prints result
func (s *Stream) Success(format string, args ...interface{}) {
	fmt.Fprintf(s, Text("success", format)+"\n", args...)
}

// Write writes the given bytes, prefixing all lines with the given prefix
func (s *Stream) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if !s.started {
		s.writer(s.Prefix)
		s.started = true
	}

	pos := bytes.IndexByte(p, '\n')
	if pos == -1 {
		s.writer(string(p))
	} else {
		pos++
		s.writer(string(p[0:pos]))
		s.started = false
		s.Write(p[pos:])
	}

	return len(p), nil
}
