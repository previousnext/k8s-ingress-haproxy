package backends

// Backends to be stored and sorted.
type Backends map[string]Backend

// Backend is a set of servers that receives forwarded requests.
type Backend struct {
	Host      string
	Path      string
	Cookie    Cookie
	Endpoints []Endpoint
}

// Cookie behaviour eg. Insert a cookie.
type Cookie struct {
	Insert bool
}

// Endpoint which is grouped in to a backend.
type Endpoint struct {
	Name string
	IP   string
	Port string
}
