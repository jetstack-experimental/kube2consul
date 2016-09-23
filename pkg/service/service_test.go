package service

import (
	"testing"

	kapi "k8s.io/kubernetes/pkg/api"
)

func TestServiceEmptyEndpointsOnePort(t *testing.T) {

	s := &Service{
		Namespace: "default",
		Name:      "one-port-service",
		k8sService: &kapi.Service{
			Spec: kapi.ServiceSpec{
				Type: kapi.ServiceTypeNodePort,
				Ports: []kapi.ServicePort{
					kapi.ServicePort{
						Name:     "http",
						NodePort: int32(9192),
						Port:     int32(80),
					},
				},
			},
		},
		k8sEndpoints: &kapi.Endpoints{},
	}

	endpoints := s.ListPorts()
	if exp, act := 1, len(endpoints); exp != act {
		t.Errorf("Element count %d is not the execpted %d", act, exp)
	}

	if exp, act := "one-port-service.default", endpoints[0].Name; exp != act {
		t.Errorf("Name '%s' is not the expected '%s'", act, exp)
	}
	if exp, act := int32(9192), endpoints[0].Port; exp != act {
		t.Errorf("Port '%d' is not the expected '%d'", act, exp)
	}
}
