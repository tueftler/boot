package command

import (
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/output"
)

type Executable interface {
	Run(stream *output.Stream) (int, error)
	String() string
}

// Boot returns a new command to be run inside a given docker container
func Boot(client *docker.Client, container *docker.Container) Executable {
	if label, ok := container.Config.Labels["boot"]; ok {
		command := strings.Split(label, " ")

		switch command[0] {
		case "NONE":
			return &None{}
		case "CMD":
			return &Exec{Client: client, Container: container, Command: append([]string{"/bin/sh", "-c"}, command[1:]...)}
		default:
			return &Exec{Client: client, Container: container, Command: command}
		}
	} else {
		return &None{}
	}
}
