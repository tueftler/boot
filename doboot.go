package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
)

const TRIES = 10

type Healthcheck struct {
	Result int
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

func healthcheck(stream *Stream, client *docker.Client, container *docker.Container) (*Healthcheck, error) {
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

	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: stream,
		ErrorStream:  stream,
		RawTerminal:  true,
	})
	if err != nil {
		return nil, err
	}

	inspect, err := client.InspectExec(exec.ID)
	if err != nil {
		return nil, err
	}

	return &Healthcheck{Result: inspect.ExitCode}, nil
}

func wait(stream *Stream, client *docker.Client, ID string) error {
	container, err := client.InspectContainer(ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(stream, "%+v\n", container.Config.Healthcheck)
	if container.Config.Healthcheck == nil || len(container.Config.Healthcheck.Test) == 0 {
		fmt.Fprintf(stream, "No Healthcheck, assuming container started\n")
		return nil
	}

	tries := TRIES
	for tries > 0 {
		check, err := healthcheck(stream, client, container)
		if err != nil {
			return err
		}

		fmt.Fprintf(stream, "Exit %d\n", check.Result)
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
