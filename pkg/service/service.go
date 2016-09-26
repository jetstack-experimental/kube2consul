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
	return fmt.Sprintf("%s.%s", s.Namespace, s.Name)
}

func (s *Service) Update() error {
	return nil
}

func (s *Service) ListPorts() []interfaces.Endpoint {
	var endpoints []interfaces.Endpoint

	portCount := len(s.k8sService.Spec.Ports)

	for _, port := range s.k8sService.Spec.Ports {
		name := fmt.Sprintf("%s-%s", s.Namespace, s.Name)
		if portCount > 1 {
			name = fmt.Sprintf("%s-%s", name, port.Name)
		}
		endpoints = append(endpoints, interfaces.Endpoint{
			Name: name,
			Port: port.NodePort,
		})
	}

	return endpoints
}

func (s *Service) ListAddresses() []string {
	addresses := make(map[string]bool)
	for _, subset := range s.k8sEndpoints.Subsets {
		for _, addr := range subset.Addresses {
			ip, err := s.kube2consul.NodeIPByPodIP(addr.IP)
			if err != nil {
				continue
			}
			addresses[ip] = true
		}
	}

	keys := make([]string, len(addresses))

	i := 0
	for k := range addresses {
		keys[i] = k
		i++
	}
	return keys
}
