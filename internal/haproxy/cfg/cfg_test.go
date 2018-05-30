package cfg

import (
	"bytes"
	"testing"

	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/backends"
	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	var buffer bytes.Buffer

	err := Generate(&buffer, GenerateParams{
		Port: 80,
		Backends: []backends.Backend{
			{
				Host: "www.example.com",
				Path: "/foo/bar",
				Cookie: backends.Cookie{
					Insert: true,
				},
				Endpoints: []backends.Endpoint{
					{
						Name: "server3",
						IP:   "3.3.3.3",
						Port: "80",
					},
				},
			},
			{
				Host: "www.example.com",
				Path: "/foo",
				Endpoints: []backends.Endpoint{
					{
						Name: "server4",
						IP:   "4.4.4.4",
						Port: "80",
					},
				},
			},
			{
				Host: "www.example.com",
				Path: "/",
				Endpoints: []backends.Endpoint{
					{
						Name: "server1",
						IP:   "1.1.1.1",
						Port: "80",
					},
					{
						Name: "server2",
						IP:   "2.2.2.2",
						Port: "80",
					},
				},
			},
		},
	})
	assert.Nil(t, err)

	assert.Contains(t, buffer.String(), "cookie SKIPPER_AFFINITY prefix")
	assert.Contains(t, buffer.String(), "cookie SKIPPER_AFFINITY insert indirect nocache")
	assert.Contains(t, buffer.String(), "server server1 1.1.1.1:80 check cookie server1")
	assert.Contains(t, buffer.String(), "server server2 2.2.2.2:80 check cookie server2")
	assert.Contains(t, buffer.String(), "server server3 3.3.3.3:80 check cookie server3")
}
