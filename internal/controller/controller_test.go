package controller

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetEndpoints(t *testing.T) {
	var (
		namespace = "example"
		name = "dev"
	)

	list := &corev1.EndpointsList{
		Items: []corev1.Endpoints{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name: "bar",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name: name,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name: "stg",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name: "prod",
				},
			},
		},
	}

	have, err := getEndpoints(list, namespace, name)
	assert.Nil(t, err)

	assert.Equal(t, have.ObjectMeta.Namespace, namespace)
	assert.Equal(t, have.ObjectMeta.Name, name)
}