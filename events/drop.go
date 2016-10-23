package events

type Drop struct {
}

// Do drops this event; clients currently subscribed will not receive
// it - as if it had never happened in the first place
func (d *Drop) Do(events *Events) {
	// NOOP
}
