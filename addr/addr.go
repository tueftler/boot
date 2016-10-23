package addr

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type Addr struct {
	Protocol string
	Network  string
	Address  string
}

// Dial opens a net.Conn
func (a *Addr) Dial() (net.Conn, error) {
	return net.Dial(a.Network, a.Address)
}

// Listen opens a net.Listener
func (a *Addr) Listen() (net.Listener, error) {
	return net.Listen(a.Network, a.Address)
}

// String returns a string representation
func (a *Addr) String() string {
	return fmt.Sprintf("%s://%s", a.Protocol, a.Address)
}

// Parse parses an input string and returns a new Addr instance. The input
// may either be a URI with the schemes "unix://", "http://" or "https://"
// or a string referring to a unix socket.
func Parse(input string) (*Addr, error) {
	pos := strings.Index(input, "://")
	if pos == -1 {
		return &Addr{Protocol: "unix", Network: "unix", Address: input}, nil
	} else {
		scheme := input[0:pos]
		pos += len("://")

		switch scheme {
		case "unix":
			return &Addr{Protocol: scheme, Network: "unix", Address: input[pos:]}, nil

		case "http", "https":
			return &Addr{Protocol: scheme, Network: "tcp", Address: input[pos:]}, nil
		}

		return nil, fmt.Errorf("Unsupported scheme '%s'", scheme)
	}
}

// Flag parses an input string. If an error occurs, prints its message, then
// runs flag.PrintDefaults() and ultimately exits the program with exitcode 1.
func Flag(input string) *Addr {
	addr, err := Parse(input)
	if err != nil {
		fmt.Print(err.Error())
		flag.PrintDefaults()
		os.Exit(1)
	}

	return addr
}
