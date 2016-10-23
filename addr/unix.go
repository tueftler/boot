package addr

import (
	"net"
	"os"
)

type UnixSocket struct {
	Path string
}

// Dial opens a net.Conn
func (u *UnixSocket) Dial() (net.Conn, error) {
	return net.Dial("unix", u.Path)
}

// Listen opens a net.Listener on this UNIX socket. If it previously
// exists, the file is removed to prevent "address already bound"
func (u *UnixSocket) Listen() (net.Listener, error) {
	os.Remove(u.Path)
	return net.Listen("unix", u.Path)
}

// String returns a string representation
func (u *UnixSocket) String() string {
	return "unix://" + u.Path
}
