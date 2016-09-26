package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/simonswine/kube2consul/pkg/mocks"
)

func TestServiceOnePort(t *testing.T) {

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

	if exp, act := "default-one-port-service", endpoints[0].Name; exp != act {
		t.Errorf("Name '%s' is not the expected '%s'", act, exp)
	}
	if exp, act := int32(9192), endpoints[0].Port; exp != act {
		t.Errorf("Port '%d' is not the expected '%d'", act, exp)
	}
}

func TestServiceTwoPorts(t *testing.T) {

	s := &Service{
		Namespace: "default",
		Name:      "two-port-service",
		k8sService: &kapi.Service{
			Spec: kapi.ServiceSpec{
				Type: kapi.ServiceTypeNodePort,
				Ports: []kapi.ServicePort{
					kapi.ServicePort{
						Name:     "http",
						NodePort: int32(9192),
						Port:     int32(80),
					},
					kapi.ServicePort{
						Name:     "https",
						NodePort: int32(9193),
						Port:     int32(443),
					},
				},
			},
		},
		k8sEndpoints: &kapi.Endpoints{},
	}

	endpoints := s.ListPorts()
	if exp, act := 2, len(endpoints); exp != act {
		t.Errorf("Element count %d is not the execpted %d", act, exp)
	}

	if exp, act := "default-two-port-service-http", endpoints[0].Name; exp != act {
		t.Errorf("Name '%s' is not the expected '%s'", act, exp)
	}
	if exp, act := int32(9192), endpoints[0].Port; exp != act {
		t.Errorf("Port '%d' is not the expected '%d'", act, exp)
	}
	if exp, act := "default-two-port-service-https", endpoints[1].Name; exp != act {
		t.Errorf("Name '%s' is not the expected '%s'", act, exp)
	}
	if exp, act := int32(9193), endpoints[1].Port; exp != act {
		t.Errorf("Port '%d' is not the expected '%d'", act, exp)
	}
}

func TestServiceNoEndpoints(t *testing.T) {

	s := &Service{
		Namespace:    "default",
		Name:         "no-endpoints",
		k8sService:   &kapi.Service{},
		k8sEndpoints: &kapi.Endpoints{},
	}

	addresses := s.ListAddresses()
	if exp, act := 0, len(addresses); exp != act {
		t.Errorf("Element count %d is not the execpted %d", act, exp)
	}
}

func TestServiceTwoEndpoints(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockK2C := mocks.NewMockKube2Consul(ctrl)
	mockK2C.EXPECT().NodeIPByPodIP("1.2.3.4").Return("192.168.0.1", nil)
	mockK2C.EXPECT().NodeIPByPodIP("4.5.6.7").Return("192.168.0.2", nil)

	s := &Service{
		Namespace:  "default",
		Name:       "no-endpoints",
		k8sService: &kapi.Service{},
		k8sEndpoints: &kapi.Endpoints{
			Subsets: []kapi.EndpointSubset{
				kapi.EndpointSubset{
					Addresses: []kapi.EndpointAddress{
						kapi.EndpointAddress{IP: "1.2.3.4"},
						kapi.EndpointAddress{IP: "4.5.6.7"},
					},
				},
			},
		},
		kube2consul: mockK2C,
	}

	addresses := s.ListAddresses()
	if exp, act := 2, len(addresses); exp != act {
		t.Errorf("Element count %d is not the execpted %d", act, exp)
	}
}

func TestServiceTwoEndpointsDuplicates(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockK2C := mocks.NewMockKube2Consul(ctrl)
	mockK2C.EXPECT().NodeIPByPodIP("1.2.3.4").Return("192.168.0.1", nil)
	mockK2C.EXPECT().NodeIPByPodIP("1.2.3.5").Return("192.168.0.1", nil)

	s := &Service{
		Namespace:  "default",
		Name:       "no-endpoints",
		k8sService: &kapi.Service{},
		k8sEndpoints: &kapi.Endpoints{
			Subsets: []kapi.EndpointSubset{
				kapi.EndpointSubset{
					Addresses: []kapi.EndpointAddress{
						kapi.EndpointAddress{IP: "1.2.3.4"},
						kapi.EndpointAddress{IP: "1.2.3.5"},
					},
				},
			},
		},
		kube2consul: mockK2C,
	}

	addresses := s.ListAddresses()
	if exp, act := 1, len(addresses); exp != act {
		t.Errorf("Element count %d is not the execpted %d", act, exp)
	}
}
