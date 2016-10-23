package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/tueftler/boot/addr"
	"github.com/tueftler/boot/output"
)

type Proxy struct {
	Forward *http.Client
	Log     *output.Stream
}

// New returns a new HTTP proxy forwarding all requests to a given address
func New(address *addr.Addr, log *output.Stream) *Proxy {
	transport := &http.Transport{Dial: func(network, addr string) (net.Conn, error) {
		return address.Dial()
	}}
	return &Proxy{Forward: &http.Client{Transport: transport}, Log: log}
}

// ServeHTTP is the http.Handler implementation
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Log.Println(">>> ", r.Method, r.URL)

	r.RequestURI = ""

	// It doesn't matter what these are set to, but they need to be set
	r.URL.Scheme = "http"
	r.URL.Host = "unix.sock"

	response, err := p.Forward.Do(r)
	if err != nil {
		p.Log.Println("<<< 502 ", err.Error())
		w.WriteHeader(502)
		fmt.Fprintf(w, "<h1>Proxy error</h1><pre>%s</pre>", err.Error())
		return
	}

	p.Log.Println("<<< ", response.Status)
	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.Header().Add("Via", "1.1 Boot")
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}
