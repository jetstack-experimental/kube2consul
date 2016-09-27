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
	for _, node := range s.ListNodes() {
		for _, port := range s.ListPorts() {
			port.NodeName = node.NodeName
			port.NodeAddress = node.NodeAddress
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
			ServiceName:      s.Name,
			ServiceNamespace: s.Namespace,
			DnsLabel:         name,
			NodePort:         port.NodePort,
		})
	}

	return endpoints
}

func (s *Service) ListNodes() []interfaces.Endpoint {
	nodes := make(map[string]bool)
	for _, subset := range s.k8sEndpoints.Subsets {
		for _, addr := range subset.Addresses {
			name, err := s.kube2consul.NodeNameByPodIP(addr.IP)
			if err != nil {
				log.Warnf("Unable to get node of PodIP %s: %s", addr.IP, err)
				continue
			}
			nodes[name] = true
		}
	}

	objects := make([]interfaces.Endpoint, len(nodes))

	i := 0
	for nodeName := range nodes {
		node, err := s.kube2consul.KubernetesClientset().Core().Nodes().Get(nodeName)
		if err != nil {
			log.Warnf("Unable to get node %s: %s", nodeName, err)
			continue
		}
		objects[i] = interfaces.Endpoint{
			NodeName:    nodeName,
			NodeAddress: node.Status.Addresses[0].Address,
		}
		i++
	}
	return objects
}
