package cfg

import (
	"io"
	"text/template"

	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/backends"
)

const tpl = `global
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

{{- range $key, $backend := .Backends }}
{{- if eq $backend.Path "/" }}
  acl {{ $key }}_domain hdr_reg(host) -i ^{{ $backend.Host }}$
  use_backend {{ $key }} if {{ $key }}_domain
{{- else }}
  acl {{ $key }}_domain hdr_reg(host) -i ^{{ $backend.Host }}$
  acl {{ $key }}_path path_sub -i {{ $backend.Path }}
  use_backend {{ $key }} if {{ $key }}_domain {{ $key }}_path
{{- end }}
{{ end }}

{{- range $key, $backend := .Backends }}
backend {{ $key }}
  balance roundrobin
  option forwardfor
  option redispatch

  cookie SKIPPER_AFFINITY prefix

{{- range $endpoint := $backend.Endpoints }}
  server {{ $endpoint.Name }} {{ $endpoint.IP }}:{{ $endpoint.Port }} check cookie {{ $endpoint.Name }}
{{- end }}
{{ end }}`

// GenerateParams passed to the Generate function.
type GenerateParams struct {
	Port     int
	Backends []backends.Backend
}

// Generate the haproxy.cfg file.
func Generate(w io.Writer, params GenerateParams) error {
	tmpl, err := template.New("haproxy").Parse(tpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, params)
}
