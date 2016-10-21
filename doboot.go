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

	if label, ok := container.Config.Labels["boot"]; ok {
		fmt.Fprintf(stream, line("info", "Using %+v"), label)

		result, err := run(stream, client, container.ID, boot(label))
		if err != nil {
			return err
		} else if result != 0 {
			return fmt.Errorf("Non-zero exit code %d", result)
		}
		return nil
	} else if container.Config.Healthcheck != nil && len(container.Config.Healthcheck.Test) > 0 {
		fmt.Fprintf(stream, line("info", "Using %+v"), container.Config.Healthcheck)

		tries := TRIES
		for tries > 0 {
			result, err := run(stream, client, container.ID, healthcheck(container.Config.Healthcheck))
			if err != nil {
				return err
			}

			fmt.Fprintf(stream, line("info", "Exit %d"), result)
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
	address := flag.String("address", "unix:///var/run/doboot.sock", "Doboot socket")

	// Docker client
	client, err := docker.NewClient(*endpoint)
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *endpoint, err.Error())
		os.Exit(1)
	}

	// HTTP proxy
	proxy, err := net.Listen("unix", "/tmp/doboot.sock")
	if err != nil {
		fmt.Printf("Error (%s) %s\n", *address, err.Error())
		os.Exit(1)
	}

	proxied := make(chan *docker.APIEvents)
	forward := &http.Client{Transport: &http.Transport{Dial: func(network, address string) (net.Conn, error) {
		return net.Dial("unix", "/var/run/docker.sock")
	}}}

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		for {
			select {
			case event := <-proxied:
				bytes, err := json.Marshal(event)
				if err != nil {
					break
				}
				w.Write(bytes)
			}
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf(">>> %s %s\n", r.Method, r.URL)
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = "unix.sock"

		response, err := forward.Do(r)
		if err != nil {
			fmt.Printf("<<< 502 %s", err.Error())
			w.WriteHeader(502)
			fmt.Fprintf(w, "<h1>Proxy error</h1><pre>%s</pre>", err.Error())
			return
		}

		fmt.Printf("<<< %s\n", response.Status)
		for header, values := range response.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", header, value)
				w.Header().Add(header, value)
			}
		}
		w.Header().Add("Via", "1.1 (DoBoot)")

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
			stream := NewStream(text("label", event.ID[0:13]+" | "), Print)
			fmt.Fprintf(stream, line("info", "%s %s (%s)"), event.Status, event.From, event.ID)

			switch event.Status {
			case "start":
				go func() {
					err := wait(stream, client, event.ID)
					if err != nil {
						fmt.Fprintf(stream, line("error", "Error %s"), err.Error())
					} else {
						fmt.Fprintf(stream, line("success", "Up and running!"))
						go func() { proxied <- event }()
					}
				}()

			default:
				go func() { proxied <- event }()
			}
		}
	}
}
