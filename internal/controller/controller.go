package controller

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy"
	"github.com/previousnext/k8s-ingress-haproxy/internal/writer"
)

// Start the new HAProxy controller.
func Start(freq time.Duration, clientset *kubernetes.Clientset, port int, file string) {
	limiter := time.Tick(freq)

	for {
		<-limiter

		fmt.Println("Collecting Ingress Rules")

		err := update(file, port, clientset)
		if err != nil {
			log.Println(err)
		}
	}
}

// Update HAProxy configuration.
func update(file string, port int, clientset *kubernetes.Clientset) error {
	h, err := haproxy.New(port)
	if err != nil {
		return errors.Wrap(err, "failed to init config builder")
	}

	ingresses, err := clientset.Extensions().Ingresses(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to lookup ingress list")
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
					return errors.Wrap(err, "failed to get endpoint list")
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
							return errors.Wrap(err, "failed to add endpoint")
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
		return errors.Wrap(err, "failed to generate configuration")
	}

	// Update it for HAProxy to consume.
	return writer.Update(b, file)
}
