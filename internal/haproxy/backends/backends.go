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
func (b Backends) Add(host, path string, endpoint Endpoint) error {
	bck := b.get(host, path)
	bck.Endpoints = append(bck.Endpoints, endpoint)
	return b.set(host, path, bck)
}

// Helper function to get a backend.
func (b Backends) get(host, path string) Backend {
	key := hash(host, path)

	if val, ok := b[key]; ok {
		return val
	}

	return Backend{
		Host: host,
		Path: path,
	}
}

// Helper function to set a backend.
func (b Backends) set(host, path string, bck Backend) error {
	key := hash(host, path)
	b[key] = bck
	return nil
}

// Helper function to create a hash based on host and path.
func hash(host, path string) string {
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(host, path)))
	return hex.EncodeToString(hasher.Sum(nil))
}
