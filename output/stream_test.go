package output

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

func Test_prefix(t *testing.T) {
	stream := NewStream("> ", func(arg string) {})
	assertEqual("> ", stream.Prefix, t)
}

func Test_prefixed(t *testing.T) {
	stream := NewStream("> ", func(arg string) {}).Prefixed("!")
	assertEqual("!", stream.Prefix, t)
}

func Test_printf(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Printf("Test %d", 0)

	assertEqual("> Test 0", written, t)
}

func Test_println(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Println("Hello", "World")

	assertEqual("> Hello World\n", written, t)
}

func Test_line(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Line("info", "Test")

	assertEqual("> "+Text("info", "Test")+"\n", written, t)
}

func Test_info(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Info("Test")

	assertEqual("> "+Text("info", "Test")+"\n", written, t)
}

func Test_error(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Error("Test")

	assertEqual("> "+Text("error", "Test")+"\n", written, t)
}

func Test_warning(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Warning("Test")

	assertEqual("> "+Text("warning", "Test")+"\n", written, t)
}

func Test_success(t *testing.T) {
	written := ""
	stream := NewStream("> ", func(arg string) { written += arg })
	stream.Success("Test")

	assertEqual("> "+Text("success", "Test")+"\n", written, t)
}
