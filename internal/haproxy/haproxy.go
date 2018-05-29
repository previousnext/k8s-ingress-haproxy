package haproxy

import (
	"fmt"
	"io"
	"text/template"

	"github.com/pkg/errors"
)

// Template for HAProxy configuration.
const Template = `global
  log /dev/log    local0
  log /dev/log    local1 notice
  chroot /var/lib/haproxy
  user haproxy
  group haproxy
  daemon

defaults
  mode    http
  log     global
  option  httplog
  option  dontlognull
  option  log-health-checks
  option  http-server-close
  timeout connect 300s
  timeout client  300s
  timeout server  300s
  errorfile 400 /etc/haproxy/errors/400.http
  errorfile 403 /etc/haproxy/errors/403.http
  errorfile 408 /etc/haproxy/errors/408.http
  errorfile 500 /etc/haproxy/errors/500.http
  errorfile 502 /etc/haproxy/errors/502.http
  errorfile 503 /etc/haproxy/errors/503.http
  errorfile 504 /etc/haproxy/errors/504.http

frontend http-in
  bind *:{{ .Port }}
  monitor-uri /ingress-status
  capture request header X-Forwarded-For len 50

{{- range $bd, $backend := .Backends }}
  acl {{ $backend.Host }} hdr_reg(host) -i ^{{ $backend.Host }}$
  acl path_root path {{ $backend.Path }}
  use_backend {{ $backend.Host }} if {{ $backend.Host }}
{{ end }}

{{- range $bd, $backend := .Backends }}
backend {{ $backend.Host }}
  balance roundrobin
  option forwardfor
  option redispatch

  cookie SKIPPER_AFFINITY prefix

{{- range $endpoint := $backend.Endpoints }}
  server {{ $endpoint.Name }} {{ $endpoint.IP }}:{{ $endpoint.Port }} check cookie {{ $endpoint.Name }}
{{- end }}
{{ end }}`

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
	tmpl, err := template.New("haproxy").Parse(Template)
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
