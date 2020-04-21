package controller

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/backends"
	"github.com/previousnext/k8s-ingress-haproxy/internal/haproxy/cfg"
	"github.com/previousnext/k8s-ingress-haproxy/internal/writer"
)

// Start the new HAProxy controller.
func Start(w io.Writer, freq time.Duration, clientset *kubernetes.Clientset, port int, file string) error {
	limiter := time.Tick(freq)

	for {
		<-limiter

		fmt.Fprintln(w,"Starting loop")

		err := update(w, file, port, clientset)
		if err != nil {
			log.Infoln(err)
		}

		fmt.Fprintln(w,"Finished")
	}

	return nil
}

// Update HAProxy configuration.
func update(w io.Writer, file string, port int, clientset *kubernetes.Clientset) error {
	bcks, err := backends.New()
	if err != nil {
		return errors.Wrap(err, "failed to init config builder")
	}

	ingresses, err := clientset.Extensions().Ingresses(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to lookup ingress list")
	}

	endpoints, err := clientset.CoreV1().Endpoints(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get Endpoints list")
	}

	if len(ingresses.Items) <= 0 {
		fmt.Fprintln(w,"No Ingress objects were found")
		return nil
	}

	// Merge Ingress -> Service -> Endpoints.
	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				endpoints, err := getEndpoints(endpoints, ingress.ObjectMeta.Namespace, path.Backend.ServiceName)
				if err != nil {
					return errors.Wrap(err, "failed to get endpoints")
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
						fmt.Fprintf(w, "Adding %s/%s to %s/%s backend list\n", address.TargetRef.Name, address.IP, rule.Host, path.Path)

						err = bcks.Add(rule.Host, path.Path, cookie, backends.Endpoint{
							Name: address.TargetRef.Name,
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
	return writer.Update(w, b, file)
}

// Helper function to get an endpoint from a list.
func getEndpoints(list *corev1.EndpointsList, namespace, name string) (corev1.Endpoints, error) {
	for _, item := range list.Items {
		if item.ObjectMeta.Namespace != namespace {
			continue
		}

		if item.ObjectMeta.Name != name {
			continue
		}

		return item, nil
	}

	return corev1.Endpoints{}, fmt.Errorf("not found: %s/%s", namespace, name)
}
