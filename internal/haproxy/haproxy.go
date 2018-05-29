package haproxy

import (
	"fmt"
	"io"

	"text/template"

	"github.com/pkg/errors"
)

// New HAProxy server.
func New(port int) (HAProxy, error) {
	haproxy := HAProxy{
		Backends: make(map[string]Backend),
	}

	if port == 0 {
		return haproxy, errors.New("port is not valid")
	}

	haproxy.Port = port

	return haproxy, nil
}

// Add an endpoint to the HAProxy configuration.
func (h HAProxy) Add(host, path string, endpoint Endpoint) error {
	bck := h.get(host, path)
	bck.Endpoints = append(bck.Endpoints, endpoint)
	return h.set(host, path, bck)
}

// Generate HAProxy configuration file.
func (h HAProxy) Generate(w io.Writer) error {
	tmpl, err := template.New("haproxy").Parse(tpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, h)
}

func (h HAProxy) get(host, path string) Backend {
	key := h.key(host, path)

	if val, ok := h.Backends[key]; ok {
		return val
	}

	return Backend{
		Host: host,
		Path: path,
	}
}

func (h HAProxy) set(host, path string, bck Backend) error {
	key := h.key(host, path)
	h.Backends[key] = bck
	return nil
}

func (h HAProxy) key(host, path string) string {
	return fmt.Sprint(host, path)
}
