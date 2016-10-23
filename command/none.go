package command

import (
	"github.com/tueftler/boot/output"
)

type None struct {
}

// Run does nothing, returns -1 as exit code
func (n *None) Run(stream *output.Stream) (int, error) {
	return -1, nil
}

// String returns a string representation of this command
func (e *None) String() string {
	return "None"
}
