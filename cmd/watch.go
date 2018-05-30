package cmd

import (
	"fmt"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/previousnext/k8s-ingress-haproxy/internal/controller"
)

type cmdWatch struct {
	Port      int
	File      string
	Frequency time.Duration
}

func (cmd *cmdWatch) run(c *kingpin.ParseContext) error {
	fmt.Println("Starting Ingress Controller")

	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to get client")
	}

	return controller.Start(cmd.Frequency, clientset, cmd.Port, cmd.File)
}

// Watch declares the "watch" sub command.
func Watch(app *kingpin.Application) {
	c := new(cmdWatch)

	cmd := app.Command("watch", "").Action(c.run)

	cmd.Flag("port", "Port to run this service on").Default("80").Envar("HAPROXY_PORT").IntVar(&c.Port)
	cmd.Flag("file", "HAProxy configuration file").Default("/etc/haproxy/haproxy.cfg").Envar("HAPROXY_CONFIG").StringVar(&c.File)
	cmd.Flag("frequency", "How frequently to recheck for new configuration").Default("5s").Envar("HAPROXY_FREQUENCY").DurationVar(&c.Frequency)
}
