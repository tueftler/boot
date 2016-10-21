package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
)

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
				fmt.Printf("> START %s: %s\n", event.From, event.ID)
			case "stop":
				fmt.Printf("> STOP %s: %s\n", event.From, event.ID)
			case "die":
				fmt.Printf("> DIE %s: %s\n", event.From, event.ID)
			}
		}
	}
}
