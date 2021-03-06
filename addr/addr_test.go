package addr

import (
	"reflect"
	"testing"
)

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func Test_unix(t *testing.T) {
	a, err := Parse("unix:///var/run/docker.sock")
	if err != nil {
		t.Error(err)
	}

	assertEqual("unix:///var/run/docker.sock", a.String(), t)
}

func Test_http(t *testing.T) {
	a, err := Parse("http://localhost")
	if err != nil {
		t.Error(err)
	}

	assertEqual("http://localhost", a.String(), t)
}

func Test_http_with_port(t *testing.T) {
	a, err := Parse("http://localhost:8080")
	if err != nil {
		t.Error(err)
	}

	assertEqual("http://localhost:8080", a.String(), t)
}

func Test_https(t *testing.T) {
	a, err := Parse("https://localhost")
	if err != nil {
		t.Error(err)
	}

	assertEqual("https://localhost", a.String(), t)
}

func Test_path(t *testing.T) {
	a, err := Parse("/var/run/docker.sock")
	if err != nil {
		t.Error(err)
	}

	assertEqual("unix:///var/run/docker.sock", a.String(), t)
}
