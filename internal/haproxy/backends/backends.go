package backends

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// New HAProxy server.
func New() (Backends, error) {
	return make(Backends), nil
}

// Add an endpoint to the HAProxy configuration.
func (h Backends) Add(host, path string, endpoint Endpoint) error {
	bck := h.get(host, path)
	bck.Endpoints = append(bck.Endpoints, endpoint)
	return h.set(host, path, bck)
}

func (b Backends) get(host, path string) Backend {
	key := b.key(host, path)

	if val, ok := b[key]; ok {
		return val
	}

	return Backend{
		Host: host,
		Path: path,
	}
}

func (b Backends) set(host, path string, bck Backend) error {
	key := b.key(host, path)
	b[key] = bck
	return nil
}

// Helper function to hash the
func (h Backends) key(host, path string) string {
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(host, path)))
	return hex.EncodeToString(hasher.Sum(nil))
}
