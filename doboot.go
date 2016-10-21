package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

const TRIES = 10

type Healthcheck struct {
	Result int
	Lines  []string
}

func writef(ID, format string, args ...interface{}) {
	fmt.Printf("\033[1;34m"+ID[0:13]+" |\033[0m "+format+"\n", args...)
}

func command(config *docker.HealthConfig) []string {
	switch config.Test[0] {
	case "CMD":
		return config.Test[1:len(config.Test)]
	case "CMD-SHELL":
		return append([]string{"/bin/sh", "-c"}, config.Test[1:len(config.Test)]...)
	default:
		return []string{"echo", "Healthcheck", config.Test[0]}
	}
}

func healthcheck(client *docker.Client, container *docker.Container) (*Healthcheck, error) {
	exec, err := client.CreateExec(docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          command(container.Config.Healthcheck),
		Container:    container.ID,
	})
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
		RawTerminal:  true,
	})
	if err != nil {
		return nil, err
	}

	inspect, err := client.InspectExec(exec.ID)
	if err != nil {
		return nil, err
	}

	return &Healthcheck{Result: inspect.ExitCode, Lines: strings.Split(stdout.String(), "\n")}, nil
}

func wait(client *docker.Client, ID string) error {
	container, err := client.InspectContainer(ID)
	if err != nil {
		return err
	}

	writef(ID, "%+v", container.Config.Healthcheck)
	if container.Config.Healthcheck == nil || len(container.Config.Healthcheck.Test) == 0 {
		writef(ID, "No Healthcheck, assuming container started")
		return nil
	}

	tries := TRIES
	for tries > 0 {
		check, err := healthcheck(client, container)
		writef(ID, "%+v", check)
		if err != nil {
			return err
		}

		if check.Result == 0 {
			return nil
		}

		time.Sleep(1 * time.Second)
		tries--
	}
	return fmt.Errorf("Timed out")
}

func main() {
	endpoint := flag.String("endpoint", "unix:///var/run/docker.sock", "Docker socket")

	client, err := docker.NewClient(*endpoint)
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *endpoint, err.Error())
		os.Exit(1)
	}

	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)
	fmt.Println("Listening...")

	for {
		select {
		case event := <-events:
			switch event.Status {
			case "start":
				writef(event.ID, "Start %s", event.From)

				go func() {
					err := wait(client, event.ID)
					if err != nil {
						writef(event.ID, "Error %s", err.Error())
					} else {
						writef(event.ID, "Up and running!")
					}
				}()

			case "stop":
				writef(event.ID, "Stop %s", event.From)
			case "die":
				writef(event.ID, "Die %s", event.From)
			}
		}
	}
}
