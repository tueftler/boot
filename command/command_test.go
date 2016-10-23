package command

import (
	"reflect"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func container(label string) *docker.Container {
	return &docker.Container{
		ID: "610036617aa165161127bc0cec60ae7831fdc1ddf1fdef1fb7f246cc83b0c315",
		Config: &docker.Config{
			Labels: map[string]string{
				"boot": label,
			},
		},
	}
}

func Test_create(t *testing.T) {
	Boot(nil, container(""))
}

func Test_command(t *testing.T) {
	fixture := Boot(nil, container("/boot.sh"))
	assertEqual([]string{"/boot.sh"}, fixture.(*Exec).Command, t)
}

func Test_none_kind(t *testing.T) {
	fixture := Boot(nil, container("NONE"))
	assertEqual("None", fixture.String(), t)
}

func Test_cmd_kind(t *testing.T) {
	fixture := Boot(nil, container("CMD /boot.sh"))
	assertEqual("Exec{[/bin/sh -c /boot.sh] @ 610036617aa16}", fixture.String(), t)
}

func Test_default_kind(t *testing.T) {
	fixture := Boot(nil, container("/boot.sh"))
	assertEqual("Exec{[/boot.sh] @ 610036617aa16}", fixture.String(), t)
}
