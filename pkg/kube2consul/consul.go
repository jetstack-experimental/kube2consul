package kube2consul

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	consulapi "github.com/hashicorp/consul/api"

	"github.com/jetstack-experimental/kube2consul/pkg/interfaces"
)

func (k *Kube2Consul) ConsulClient() *consulapi.Client {
	if k.consulClient == nil {
		config := consulapi.DefaultConfig()
		config.Address = k.consulAddress
		client, err := consulapi.NewClient(config)
		if err != nil {
			panic(err.Error())
		}
		k.consulClient = client
	}
	return k.consulClient
}

func (k *Kube2Consul) ConsulCatalog() *consulapi.Catalog {
	if k.consulClient == nil {
		k.consulCatalog = k.ConsulClient().Catalog()
	}
	return k.consulCatalog
}

func (k *Kube2Consul) UpdateConsul(namespace string, name string, endpoints []interfaces.Endpoint) error {

	tag := fmt.Sprintf("kube2consul-%s/%s", namespace, name)
	// TODO Get existing services and remove them

	for _, endpoint := range endpoints {
		service := &consulapi.AgentService{
			Service: endpoint.DnsLabel,
			Tags:    []string{tag},
			Port:    int(endpoint.NodePort),
		}
		reg := &consulapi.CatalogRegistration{
			Node:    endpoint.NodeName,
			Address: endpoint.NodeAddress,
			Service: service,
		}
		_, err := k.ConsulCatalog().Register(reg, &consulapi.WriteOptions{})
		if err != nil {
			log.Warnf("Error registering %+v: %s", reg, err)
		}
	}
	return nil
}
