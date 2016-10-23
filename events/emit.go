package events

import (
	"github.com/fsouza/go-dockerclient"
)

type Emit struct {
	Event *docker.APIEvents
}

// Do emits the event, causing it to be sent on to all clients currently
// listening for Docker API events
func (e *Emit) Do(events *Events) {
	events.Emit(e.Event)
}
