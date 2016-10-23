package addr

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type Addr interface {
	Dial() (net.Conn, error)
	Listen() (net.Listener, error)
	String() string
}

// Parse parses an input string and returns a new Addr instance. The input
// may either be a URI with the schemes "unix://", "http://" or "https://"
// or a string referring to a unix socket.
func Parse(input string) (Addr, error) {
	pos := strings.Index(input, "://")
	if pos == -1 {
		return &UnixSocket{Path: input}, nil
	} else {
		scheme := input[0:pos]
		pos += len("://")

		switch scheme {
		case "unix":
			return &UnixSocket{Path: input[pos:]}, nil

		case "http", "https":
			return &HttpEndpoint{Scheme: scheme, Host: input[pos:]}, nil
		}

		return nil, fmt.Errorf("Unsupported scheme '%s'", scheme)
	}
}

// Flag parses an input string. If an error occurs, prints its message, then
// runs flag.PrintDefaults() and ultimately exits the program with exitcode 1.
func Flag(input string) Addr {
	addr, err := Parse(input)
	if err != nil {
		fmt.Print(err.Error())
		flag.PrintDefaults()
		os.Exit(1)
	}

	return addr
}
