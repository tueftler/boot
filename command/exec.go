package command

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/output"
)

type Exec struct {
	Client    *docker.Client
	Container string
	Command   []string
}

// Run executes the given command line inside the container, streaming
// it STDOUT and STDERR to the given stream and returning its exitcode.
func (e *Exec) Run(stream *output.Stream) (int, error) {
	exec, err := e.Client.CreateExec(docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          e.Command,
		Container:    e.Container,
	})
	if err != nil {
		return -1, err
	}

	err = e.Client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: stream,
		ErrorStream:  stream,
		RawTerminal:  false,
	})
	if err != nil {
		return -1, err
	}

	inspect, err := e.Client.InspectExec(exec.ID)
	if err != nil {
		return -1, err
	}

	return inspect.ExitCode, nil
}

// String returns a string representation of this command
func (e *Exec) String() string {
	return fmt.Sprintf("Exec{%s @ %s}", e.Command, e.Container[0:13])
}
