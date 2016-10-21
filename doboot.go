package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

const TRIES = 10

func boot(label string) []string {
	command := strings.Split(label, " ")
	switch command[0] {
	case "CMD":
		return append([]string{"/bin/sh", "-c"}, command[1:len(command)]...)
	default:
		return command
	}
}

func healthcheck(config *docker.HealthConfig) []string {
	switch config.Test[0] {
	case "CMD":
		return config.Test[1:len(config.Test)]
	case "CMD-SHELL":
		return append([]string{"/bin/sh", "-c"}, config.Test[1:len(config.Test)]...)
	default:
		return []string{"echo", "Healthcheck", config.Test[0]}
	}
}

func run(stream *Stream, client *docker.Client, id string, cmd []string) (int, error) {
	exec, err := client.CreateExec(docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          cmd,
		Container:    id,
	})
	if err != nil {
		return -1, err
	}

	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: stream,
		ErrorStream:  stream,
		RawTerminal:  true,
	})
	if err != nil {
		return -1, err
	}

	inspect, err := client.InspectExec(exec.ID)
	if err != nil {
		return -1, err
	}

	return inspect.ExitCode, nil
}

func wait(stream *Stream, client *docker.Client, ID string) error {
	container, err := client.InspectContainer(ID)
	if err != nil {
		return err
	}

	if label, ok:= container.Config.Labels["boot"]; ok {
		fmt.Fprintf(stream, "Using %+v\n", label)

		result, err := run(stream, client, container.ID, boot(label))
		if err != nil {
			return err
		} else if result != 0 {
			return fmt.Errorf("Non-zero exit code %d", result)
		}
		return nil
	} else if container.Config.Healthcheck != nil && len(container.Config.Healthcheck.Test) > 0 {
		fmt.Fprintf(stream, "Using %+v\n", container.Config.Healthcheck)

		tries := TRIES
		for tries > 0 {
			result, err := run(stream, client, container.ID, healthcheck(container.Config.Healthcheck))
			if err != nil {
				return err
			}

			fmt.Fprintf(stream, "Exit %d\n", result)
			if result == 0 {
				return nil
			}

			time.Sleep(1 * time.Second)
			tries--
		}
		return fmt.Errorf("Timed out")
	} else {		
		fmt.Fprintf(stream, "Neither boot command nor healthcheck present, assuming container started\n")
		return nil
	}
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
			stream := NewStream("\033[1;34m"+event.ID[0:13]+" |\033[0m ", Print)
			switch event.Status {
			case "start":
				fmt.Fprintf(stream, "Start %s\n", event.From)

				go func() {
					err := wait(stream, client, event.ID)
					if err != nil {
						fmt.Fprintf(stream, "Error %s\n", err.Error())
					} else {
						fmt.Fprintf(stream, "Up and running!\n")
					}
				}()

			case "stop":
				fmt.Fprintf(stream, "Stop %s\n", event.From)
			case "die":
				fmt.Fprintf(stream, "Die %s\n", event.From)
			}
		}
	}
}
