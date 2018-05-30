package main

import (
	"os"

	"github.com/alecthomas/kingpin"

	"github.com/previousnext/k8s-ingress-haproxy/cmd"
)

func main() {
	app := kingpin.New("k8s-ingress-haproxy", "Ingress Controller for HAProxy")

	cmd.Version(app)
	cmd.Watch(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
