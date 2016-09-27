package service

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/simonswine/kube2consul/pkg/interfaces"
)

type Service struct {
	Namespace    string
	Name         string
	kube2consul  interfaces.Kube2Consul
	k8sService   *kapi.Service
	k8sEndpoints *kapi.Endpoints
	mutex        sync.Mutex
	TestString   string
}

func New(kube2consul interfaces.Kube2Consul, namespace string, name string) *Service {
	svc := Service{
		Name:        name,
		Namespace:   namespace,
		kube2consul: kube2consul,
	}
	return &svc
}

func (s *Service) Key() string {
	return fmt.Sprintf("%s.%s", s.Namespace, s.Name)
}

func (s *Service) Update() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// object not filled
	if s.k8sEndpoints == nil || s.k8sService == nil {
		return nil
	}

	// only look at NodePort services
	if s.k8sService.Spec.Type != kapi.ServiceTypeNodePort {
		return nil
	}

	list := s.List()
	log.Debugf("Endpoints %+v", list)

	return nil
}

func (s *Service) UpdateEndpoints(endpoints *kapi.Endpoints) error {
	s.mutex.Lock()
	s.k8sEndpoints = endpoints
	s.mutex.Unlock()
	return nil
}

func (s *Service) UpdateService(svc *kapi.Service) error {
	s.mutex.Lock()
	s.k8sService = svc
	s.mutex.Unlock()
	return nil
}
func (s *Service) List() []interfaces.Endpoint {
	var endpoints []interfaces.Endpoint
	for _, address := range s.ListAddresses() {
		for _, port := range s.ListPorts() {
			port.Address = address
			endpoints = append(endpoints, port)
		}
	}
	return endpoints
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
