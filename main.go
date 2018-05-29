package main

import (
	"fmt"

	"github.com/alecthomas/kingpin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/previousnext/k8s-ingress-haproxy/internal/controller"
)

var (
	cliPort      = kingpin.Flag("port", "Port to run this service on").Default("80").Envar("HAPROXY_PORT").Int()
	cliFile      = kingpin.Flag("config", "HAProxy configuration file").Default("/etc/haproxy/haproxy.cfg").Envar("HAPROXY_CONFIG").String()
	cliFrequency = kingpin.Flag("frequency", "How frequently to recheck for new configuration").Default("5s").Envar("HAPROXY_FREQUENCY").Duration()
)

func main() {
	kingpin.Parse()

	fmt.Println("Starting Ingress Controller")

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	controller.Start(*cliFrequency, clientset, *cliPort, *cliFile)
}
