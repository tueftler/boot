package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/addr"
	"github.com/tueftler/boot/command"
	"github.com/tueftler/boot/events"
	"github.com/tueftler/boot/output"
	"github.com/tueftler/boot/proxy"
)

func start(log *output.Stream, client *docker.Client, event *docker.APIEvents) events.Action {
	stream := log.Prefixed(output.Text("container", event.Actor.ID[0:13]+" | "))

	container, err := client.InspectContainer(event.Actor.ID)
	if err != nil {
		stream.Error("Inspect error %s", err.Error())
		return &events.Drop{}
	}

	if label, ok := container.Config.Labels["boot"]; ok {
		boot := command.Boot(client, container.ID, label)
		stream.Info("Using %s", boot)

		result, err := boot.Run(stream)
		if err != nil {
			stream.Error("Run error %s", err.Error())
			return &events.Drop{}
		} else if result != 0 {
			stream.Error("Non-zero exit code %d", result)
			return &events.Drop{}
		}

		stream.Success("Up and running!")
		return &events.Emit{Event: event}
	} else {
		stream.Warning("No boot command present, assuming container started")
		return &events.Emit{Event: event}
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

	// Distribute events
	events := events.Distribute(client, output.NewStream(output.Text("proxy", "distribute    | "), output.Print))
	http.Handle("/events", events)
	http.Handle("/v1.24/events", events)
	http.Handle("/v1.19/events", events)
	http.Handle("/v1.12/events", events)

	// Proxy the rest of the API calls
	http.Handle("/", proxy.Pass(endpoint, output.NewStream(output.Text("proxy", "proxy         | "), output.Print)))

	go http.Serve(listen, nil)

	events.Intercept("start", start)
	events.Log.Info("Listening...")
	events.Listen()
}
