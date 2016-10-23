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

	boot := command.Boot(client, container)
	stream.Info("Using boot command %s", boot)
	result, err := boot.Run(stream)
	if err != nil {
		stream.Error("Run error %s", err.Error())
		return &events.Drop{}
	}

	switch result {
	case -1:
		stream.Warning("No boot command present, assuming container started")
		return &events.Emit{Event: event}

	case 0:
		stream.Success("Up and running!")
		return &events.Emit{Event: event}

	default:
		stream.Error("Non-zero exit code %d", result)
		return &events.Drop{}
	}
}

func run(connect, listen *addr.Addr) error {
	client, err := docker.NewClient(connect.String())
	if err != nil {
		return fmt.Errorf("%s: %s", connect, err.Error())
	}

	server, err := listen.Listen()
	if err != nil {
		return fmt.Errorf("%s: %s", listen, err.Error())
	}

	events := events.Distribute(client, output.NewStream(output.Text("proxy", "distribute    | "), output.Print))
	proxy := proxy.Pass(connect, output.NewStream(output.Text("proxy", "proxy         | "), output.Print))

	urls := http.NewServeMux()
	urls.Handle("/events", events)
	urls.Handle("/v1.24/events", events)
	urls.Handle("/v1.19/events", events)
	urls.Handle("/v1.12/events", events)
	urls.Handle("/", proxy)

	go http.Serve(server, urls)

	events.Intercept("start", start)
	events.Log.Info("Listening...")
	events.Listen()

	return nil
}

func main() {
	docker := flag.String("docker", "unix:///var/run/docker.sock", "Docker socket")
	listen := flag.String("listen", "unix:///var/run/boot.sock", "Boot socket")
	flag.Parse()

	if err := run(addr.Flag(*docker), addr.Flag(*listen)); err != nil {
		fmt.Printf("Error %s\n", err.Error())
		os.Exit(1)
	}
}
