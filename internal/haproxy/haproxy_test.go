package haproxy

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	h, err := New(80)
	assert.Nil(t, err)

	h.Add("www.example.com", "/", Endpoint{
		Name: "server1",
		IP:   "1.1.1.1",
		Port: "80",
	})

	h.Add("www.example.com", "/", Endpoint{
		Name: "server2",
		IP:   "2.2.2.2",
		Port: "80",
	})

	h.Add("www.example.com", "/foo", Endpoint{
		Name: "server3",
		IP:   "3.3.3.3",
		Port: "80",
	})

	h.Add("www.example.com", "/foo", Endpoint{
		Name: "server4",
		IP:   "4.4.4.4",
		Port: "80",
	})

	var buffer bytes.Buffer

	err = h.Generate(&buffer)
	assert.Nil(t, err)

	assert.Contains(t, buffer.String(), "server server1 1.1.1.1:80 check cookie server1")
	assert.Contains(t, buffer.String(), "server server2 2.2.2.2:80 check cookie server2")
	assert.Contains(t, buffer.String(), "server server3 3.3.3.3:80 check cookie server3")
}
