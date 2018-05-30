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

	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/backends"
	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/cfg"
	"github.com/previousnext/k8s-ingress-haproxy/internal/writer"
)

// Start the new HAProxy controller.
func Start(freq time.Duration, clientset *kubernetes.Clientset, port int, file string) error {
	limiter := time.Tick(freq)

	for {
		<-limiter

		fmt.Println("Collecting Ingress Rules")

		err := update(file, port, clientset)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

// Update HAProxy configuration.
func update(file string, port int, clientset *kubernetes.Clientset) error {
	bcks, err := backends.New()
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

				// Don't insert a cookie by default.
				cookie := false

				// @todo, Update the annotation to follow Ingress conventions:
				// https://docs.traefik.io/configuration/backends/kubernetes/#general-annotations
				if _, ok := ingress.ObjectMeta.Annotations["cookieInsert"]; ok {
					// Warning! This breaks cache.
					cookie = true
				}

				for _, subnet := range endpoints.Subsets {
					for _, address := range subnet.Addresses {
						err = bcks.Add(rule.Host, path.Path, cookie, backends.Endpoint{
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
	err = cfg.Generate(&b, cfg.GenerateParams{
		Port:     port,
		Backends: bcks.Sorted(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to generate configuration")
	}

	// Update it for HAProxy to consume.
	return writer.Update(b, file)
}
