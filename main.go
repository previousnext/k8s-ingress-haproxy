package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"
	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	cliPort      = kingpin.Flag("port", "Port to run this service on").Default("80").Envar("HAPROXY_PORT").Int()
	cliFile      = kingpin.Flag("config", "HAProxy configuration file").Default("/etc/haproxy/haproxy.cfg").Envar("HAPROXY_CONFIG").String()
	cliFrequency = kingpin.Flag("frequency", "How frequently to recheck for new configuration").Default("5s").Envar("HAPROXY_FREQUENCY").Duration()
)

func main() {
	kingpin.Parse()

	fmt.Println("Starting Ingress Controller")

	limiter := time.Tick(*cliFrequency)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for {
		<-limiter

		fmt.Println("Collecting Ingress Rules")

		err := generate(*cliFile, *cliPort, clientset)
		if err != nil {
			log.Println(err)
		}
	}
}

// Generate the new configuration file.
func generate(file string, port int, clientset *kubernetes.Clientset) error {
	h, err := haproxy.New(*cliPort)
	if err != nil {
		return errors.Wrap(err, "failed to init HAProxy config builder")
	}

	ingresses, err := clientset.Extensions().Ingresses(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to lookup Ingress list")
	}

	if len(ingresses.Items) <= 0 {
		return errors.New("no Ingress objects were found")
	}

	// Merge Ingress -> Service -> Endpoints.
	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				endpoints, err := clientset.CoreV1().Endpoints(ingress.ObjectMeta.Namespace).Get(path.Backend.ServiceName, metav1.GetOptions{})
				if err != nil {
					return errors.Wrap(err, "failed to get Endpoint list")
				}

				for _, subnet := range endpoints.Subsets {
					for _, address := range subnet.Addresses {
						err = h.Add(rule.Host, path.Path, haproxy.Endpoint{
							Name: address.Hostname,
							IP:   address.IP,
							// @todo, Remove hardcoded value.
							Port: "80",
						})
						if err != nil {
							return errors.Wrap(err, "failed to add Endpoint to HAProxy configuration")
						}
					}
				}
			}
		}
	}

	var b bytes.Buffer

	// Generate the HAProxy configuration file.
	err = h.Generate(&b)
	if err != nil {
		return nil
	}

	// Update it for HAProxy to consume.
	return update(b, file)
}

// Update if the file has changed.
func update(update bytes.Buffer, file string) error {
	// Check if the file has changed, if not, lets create it for the first time.
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return write(update, file)
	}

	// Load the existing config file so we can compare with the new.
	existing, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "failed to read existing file")
	}

	// Is this the same file?
	if update.String() == string(existing) {
		return errors.New("file has not changed")
	}

	// It is not, lets write update it.
	return write(update, file)
}

// Write the configuration file for HAProxy to consume.
func write(update bytes.Buffer, file string) error {
	// Create a new file which we can apply our template to.
	w, err := os.Create(file)
	if err != nil {
		return err
	}

	// Write to the file.
	_, err = w.Write(update.Bytes())
	if err != nil {
		return err
	}

	return nil
}
