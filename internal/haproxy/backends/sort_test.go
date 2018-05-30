package backends

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	b, err := New()
	assert.Nil(t, err)

	b.Add("www.example.com", "/", Endpoint{
		Name: "server1",
		IP:   "1.1.1.1",
		Port: "80",
	})

	b.Add("www.example.com", "/", Endpoint{
		Name: "server2",
		IP:   "2.2.2.2",
		Port: "80",
	})

	b.Add("www.example.com", "/foo/bar", Endpoint{
		Name: "server3",
		IP:   "3.3.3.3",
		Port: "80",
	})

	b.Add("www.example.com", "/foo", Endpoint{
		Name: "server4",
		IP:   "4.4.4.4",
		Port: "80",
	})

	expected := []Backend{
		{
			Host: "www.example.com",
			Path: "/foo/bar",
			Endpoints: []Endpoint{
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
			Endpoints: []Endpoint{
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
			Endpoints: []Endpoint{
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
	}

	assert.Equal(t, expected, b.Sorted())
}
