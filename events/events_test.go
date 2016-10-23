package events

import (
	"reflect"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/tueftler/boot/output"
)

const CONTAINER = "610036617aa165161127bc0cec60ae7831fdc1ddf1fdef1fb7f246cc83b0c315"

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func Test_create(t *testing.T) {
	Distribute(nil, output.NewStream("", output.Print))
}

func Test_handle(t *testing.T) {
	written := ""
	log := output.NewStream("> ", func(arg string) { written += arg })

	fixture := Distribute(nil, log)
	fixture.Handle(&docker.APIEvents{Action: "start", Actor: docker.APIActor{ID: CONTAINER}})

	assertEqual("> To 0 -> start 610036617aa16 map[]\n", written, t)
}

func Test_handle_emit(t *testing.T) {
	written := ""
	log := output.NewStream("> ", func(arg string) { written += arg })

	fixture := Distribute(nil, log)
	fixture.Intercept("start", func(log *output.Stream, client *docker.Client, event *docker.APIEvents) Action {
		return &Emit{Event: event}
	})
	fixture.Handle(&docker.APIEvents{Action: "start", Actor: docker.APIActor{ID: CONTAINER}})

	assertEqual("> To 0 -> start 610036617aa16 map[]\n", written, t)
}

func Test_handle_stream_writes(t *testing.T) {
	written := ""
	log := output.NewStream("> ", func(arg string) { written += arg })

	fixture := Distribute(nil, log)
	fixture.Intercept("start", func(log *output.Stream, client *docker.Client, event *docker.APIEvents) Action {
		log.Println("Handling")
		return &Emit{Event: event}
	})
	fixture.Handle(&docker.APIEvents{Action: "start", Actor: docker.APIActor{ID: CONTAINER}})

	assertEqual("> Handling\n> To 0 -> start 610036617aa16 map[]\n", written, t)
}

func Test_handle_drop(t *testing.T) {
	written := ""
	log := output.NewStream("> ", func(arg string) { written += arg })

	fixture := Distribute(nil, log)
	fixture.Intercept("start", func(log *output.Stream, client *docker.Client, event *docker.APIEvents) Action {
		return &Drop{}
	})
	fixture.Handle(&docker.APIEvents{Action: "start", Actor: docker.APIActor{ID: CONTAINER}})

	assertEqual("", written, t)
}
