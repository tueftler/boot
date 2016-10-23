package command

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/output"
)

type Command struct {
	Client    *docker.Client
	Container string
	Command   []string
}

// New returns a new command to be run inside a given docker container
func New(client *docker.Client, container string, command []string) *Command {
	return &Command{Client: client, Container: container, Command: command}
}

// Run runs the command, streaming its output tot the given stream and
// returning its exitcode.
func (c *Command) Run(stream *output.Stream) (int, error) {
	exec, err := c.Client.CreateExec(docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          c.Command,
		Container:    c.Container,
	})
	if err != nil {
		return -1, err
	}

	err = c.Client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: stream,
		ErrorStream:  stream,
		RawTerminal:  false,
	})
	if err != nil {
		return -1, err
	}

	inspect, err := c.Client.InspectExec(exec.ID)
	if err != nil {
		return -1, err
	}

	return inspect.ExitCode, nil
}

// String returns a string representation of this command
func (c *Command) String() string {
	return fmt.Sprintf("%s @ %s", c.Command, c.Container[0:13])
}