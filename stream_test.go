package main

import (
	"io"
	"reflect"
	"testing"
)

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func Test_create(t *testing.T) {
	NewStream("> ", Print)
}

func Test_writing(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "Test")

	assertEqual("> Test", written, t)
}

func Test_writing_newline(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "\n")

	assertEqual("> \n", written, t)
}

func Test_multiple_writes(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "T")
	io.WriteString(stream, "e")
	io.WriteString(stream, "s")
	io.WriteString(stream, "t")
	io.WriteString(stream, "\n")

	assertEqual("> Test\n", written, t)
}

func Test_writing_chunk_with_newline(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "T\na")

	assertEqual("> T\n> a", written, t)
}

func Test_writing_line(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "Test\n")

	assertEqual("> Test\n", written, t)
}

func Test_writing_lines(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	io.WriteString(stream, "Line 1\n")
	io.WriteString(stream, "Line 2\n")

	assertEqual("> Line 1\n> Line 2\n", written, t)
}
