package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/addr"
	"github.com/tueftler/boot/output"
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
	} else if container.Config.Healthcheck != nil && len(container.Config.Healthcheck.Test) > 0 {
		stream.Line("info", "Using %+v", container.Config.Healthcheck)

		tries := TRIES
		for tries > 0 {
			result, err := run(stream, client, container.ID, healthcheck(container.Config.Healthcheck))
			if err != nil {
				return err
			}

			stream.Line("info", "Exit %d", result)
			if result == 0 {
				return nil
			}

			time.Sleep(1 * time.Second)
			tries--
		}
		return fmt.Errorf("Timed out")
	} else {
		stream.Line("warning", "Neither boot command nor healthcheck present, assuming container started")
		return nil
	}
}

func distribute(stream *output.Stream, listeners []chan *docker.APIEvents, event *docker.APIEvents) {
	for _, listener := range listeners {
		listener <- event
	}

	stream.Printf("To %d -> %s %s %+v\n", len(listeners), event.Action, event.Actor.ID[0:13], event.Actor.Attributes)
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
	proxy, err := addr.Flag(*listenSocket).Listen()
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *listenSocket, err.Error())
		os.Exit(1)
	}

	var listeners []chan *docker.APIEvents

	// Force all forwarded traffic to docker socket
	forward := &http.Client{Transport: &http.Transport{Dial: func(network, address string) (net.Conn, error) {
		return endpoint.Dial()
	}}}

	handle := func(w http.ResponseWriter, r *http.Request) {

		// TODO Use http://stackoverflow.com/a/18897083
		listener := make(chan *docker.APIEvents)
		listeners = append(listeners, listener)
		index := len(listeners)

		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)

		if f, ok := w.(http.Flusher); ok {
			for {
				select {
				case event := <-listener:
					bytes, _ := json.Marshal(event)
					if err != nil {
						panic(err)
					}

					if _, err := w.Write(bytes); err != nil {
						listeners = append(listeners[:index-1], listeners[index:]...)
						return
					}

					f.Flush()
				}
			}
		}
	}
	http.HandleFunc("/events", handle)
	http.HandleFunc("/v1.24/events", handle)
	http.HandleFunc("/v1.19/events", handle)
	http.HandleFunc("/v1.12/events", handle)

	stream := output.NewStream(output.Text("proxy", "proxy         | "), output.Print)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stream.Println(">>> ", r.Method, r.URL)
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = "unix.sock"

		response, err := forward.Do(r)
		if err != nil {
			stream.Println("<<< 502 ", err.Error())
			w.WriteHeader(502)
			fmt.Fprintf(w, "<h1>Proxy error</h1><pre>%s</pre>", err.Error())
			return
		}

		stream.Println("<<< ", response.Status)
		for header, values := range response.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}
		w.Header().Add("Via", "1.1 Boot")

		w.WriteHeader(response.StatusCode)
		io.Copy(w, response.Body)
	})

	go http.Serve(proxy, nil)

	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)
	fmt.Println("Listening...")

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
