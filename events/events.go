package events

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/output"
)

type Action interface {
	Do(events *Events)
}

type Handler func(stream *output.Stream, client *docker.Client, event *docker.APIEvents) Action

type Events struct {
	Client    *docker.Client
	Log       *output.Stream
	Listeners []chan *docker.APIEvents
	Handlers  map[string]Handler
}

// Distribute returns an events instance which is able to distribute
// received Docker API events to listeners, optionally intercepting them.
func Distribute(client *docker.Client, stream *output.Stream) *Events {
	return &Events{
		Client:    client,
		Log:       stream,
		Listeners: make([]chan *docker.APIEvents, 0),
		Handlers:  make(map[string]Handler),
	}
}

// Emit distributes an event to all listeners and logs it
func (e *Events) Emit(event *docker.APIEvents) {
	for _, listener := range e.Listeners {
		listener <- event
	}
	e.Log.Printf("To %d -> %s %s %+v\n", len(e.Listeners), event.Action, event.Actor.ID[0:13], event.Actor.Attributes)
}

// Intercept adds a handler for intercepting a given named event
func (e *Events) Intercept(name string, handler Handler) {
	e.Handlers[name] = handler
}

// Handle handles a single event
func (e *Events) Handle(event *docker.APIEvents) {
	if handler, ok := e.Handlers[event.Action]; ok {
		handler(e.Log, e.Client, event).Do(e)
	} else {
		e.Emit(event)
	}
}

// Listen starts listening for events on the Docker API and passes them
// to Handle() when they occur, not waiting for it to return.
func (e *Events) Listen() {
	events := make(chan *docker.APIEvents)
	e.Client.AddEventListener(events)
	defer e.Client.RemoveEventListener(events)

	for {
		select {
		case event := <-events:
			go e.Handle(event)
		}
	}
}

// ServeHTTP is the http.Handler implementation
func (e *Events) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var l sync.Mutex

	listener := make(chan *docker.APIEvents)

	l.Lock()
	e.Listeners = append(e.Listeners, listener)
	index := len(e.Listeners)
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
					e.Listeners = append(e.Listeners[:index-1], e.Listeners[index:]...)
					l.Unlock()
					return
				}

				f.Flush()
			}
		}
	}
}
