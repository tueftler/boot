package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/addr"
	"github.com/tueftler/boot/output"
	"github.com/tueftler/boot/proxy"
)

const PROXY = "proxy         | "
const DISTRIBUTE = "distribute    | "

var listeners []chan *docker.APIEvents

func boot(label string) []string {
	command := strings.Split(label, " ")
	switch command[0] {
	case "CMD":
		return append([]string{"/bin/sh", "-c"}, command[1:]...)
	default:
		return command
	}
}

func run(stream *output.Stream, client *docker.Client, id string, cmd []string) (int, error) {
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
		RawTerminal:  false,
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

func wait(stream *output.Stream, client *docker.Client, ID string) error {
	container, err := client.InspectContainer(ID)
	if err != nil {
		return err
	}

	if label, ok := container.Config.Labels["boot"]; ok {
		stream.Line("info", "Using %+v", label)

		result, err := run(stream, client, container.ID, boot(label))
		if err != nil {
			return err
		} else if result != 0 {
			return fmt.Errorf("Non-zero exit code %d", result)
		}
		return nil
	} else {
		stream.Line("warning", "No boot command present, assuming container started")
		return nil
	}
}

func distribute(stream *output.Stream, listeners []chan *docker.APIEvents, event *docker.APIEvents) {
	for _, listener := range listeners {
		listener <- event
	}

	stream.Printf("To %d -> %s %s %+v\n", len(listeners), event.Action, event.Actor.ID[0:13], event.Actor.Attributes)
}

func events(w http.ResponseWriter, r *http.Request) {
	var l sync.Mutex

	listener := make(chan *docker.APIEvents)

	l.Lock()
	listeners = append(listeners, listener)
	index := len(listeners)
	l.Unlock()

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	if f, ok := w.(http.Flusher); ok {
		for {
			select {
			case event := <-listener:
				bytes, _ := json.Marshal(event)
				if _, err := w.Write(bytes); err != nil {
					l.Lock()
					listeners = append(listeners[:index-1], listeners[index:]...)
					l.Unlock()
					return
				}

				f.Flush()
			}
		}
	}
}

func main() {
	dockerSocket := flag.String("docker", "unix:///var/run/docker.sock", "Docker socket")
	listenSocket := flag.String("listen", "unix:///var/run/boot.sock", "Boot socket")
	flag.Parse()

	// Docker client
	endpoint := addr.Flag(*dockerSocket)
	client, err := docker.NewClient(endpoint.String())
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *dockerSocket, err.Error())
		os.Exit(1)
	}

	// HTTP proxy
	listen, err := addr.Flag(*listenSocket).Listen()
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *listenSocket, err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/events", events)
	http.HandleFunc("/v1.24/events", events)
	http.HandleFunc("/v1.19/events", events)
	http.HandleFunc("/v1.12/events", events)
	http.Handle("/", proxy.New(endpoint, output.NewStream(output.Text("proxy", PROXY), output.Print)))
	go http.Serve(listen, nil)

	// Listen for events
	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)

	stream := output.NewStream(output.Text("proxy", DISTRIBUTE), output.Print)
	stream.Line("info", "Listening...")
	for {
		select {
		case event := <-events:
			switch event.Status {
			case "start":
				stream := output.NewStream(output.Text("docker", event.ID[0:13]+" | "), output.Print)
				go func() {
					err := wait(stream, client, event.ID)
					if err != nil {
						stream.Line("error", "Error %s", err.Error())
					} else {
						stream.Line("success", "Up and running!")
						distribute(stream, listeners, event)
					}
				}()

			default:
				distribute(stream, listeners, event)
			}
		}
	}
}
