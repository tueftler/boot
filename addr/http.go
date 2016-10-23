package addr

import "net"

type HttpEndpoint struct {
	Scheme string
	Host   string
}

// Dial opens a net.Conn
func (h *HttpEndpoint) Dial() (net.Conn, error) {
	return net.Dial("tcp", h.Host)
}

// Listen opens a net.Listener
func (h *HttpEndpoint) Listen() (net.Listener, error) {
	return net.Listen("tcp", h.Host)
}

// String returns a string representation
func (h *HttpEndpoint) String() string {
	return h.Scheme + "://" + h.Host
}
