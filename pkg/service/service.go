package service

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/simonswine/kube2consul/pkg/interfaces"
)

type Service struct {
	Namespace    string
	Name         string
	kube2consul  interfaces.Kube2Consul
	k8sService   *kapi.Service
	k8sEndpoints *kapi.Endpoints
}

func (s *Service) Key() string {
	return fmt.Sprintf("%s.%s", s.Name, s.Namespace)
}

func (s *Service) Update() error {
	return nil
}

func (s *Service) ListPorts() []interfaces.Endpoint {
	var endpoints []interfaces.Endpoint

	portCount := len(s.k8sService.Spec.Ports)

	for _, port := range s.k8sService.Spec.Ports {
		name := s.Key()
		if portCount > 1 {
			name = fmt.Sprintf("%s.%s", port.Name, name)
		}
		endpoints = append(endpoints, interfaces.Endpoint{
			Name: name,
			Port: port.NodePort,
		})
	}

	return endpoints
}
