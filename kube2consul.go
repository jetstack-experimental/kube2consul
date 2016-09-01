package main

import (
	"log"

	kapi "k8s.io/kubernetes/pkg/api"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"

	consulapi "github.com/hashicorp/consul/api"
)

type kube2consul struct {
	kubeClient    *kclient.Client
	consulClient  *consulapi.Client
	consulCatalog *consulapi.Catalog
}

func newKube2Consul(kc *kclient.Client, cc *consulapi.Client) *kube2consul {
	k2c := &kube2consul{
		kubeClient:    kc,
		consulClient:  cc,
		consulCatalog: cc.Catalog(),
	}
	return k2c
}

// watchForServices starts watching for new, removed or updated kubernetes services
func (kc *kube2consul) watchForServices() kcache.Store {
	serviceStore, serviceController := kframework.NewInformer(
		kcache.NewListWatchFromClient(kc.kubeClient, "services", kapi.NamespaceAll, kselector.Everything()),
		&kapi.Service{},
		resyncPeriod,
		kframework.ResourceEventHandlerFuncs{
			AddFunc:    kc.newService,
			DeleteFunc: kc.removeService,
			UpdateFunc: kc.updateService,
		},
	)
	go serviceController.Run(wait.NeverStop)
	return serviceStore
}

// newService registers a new kubernetes service in Consul
func (kc *kube2consul) newService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		log.Printf("Add Service %+v\n", s.GetName())
		service := &consulapi.AgentService{
			Service: s.GetName(),
			Tags:    []string{"kubernetes"},
		}
		if len(s.Spec.Ports) > 0 {
			service.Port = int(s.Spec.Ports[0].Port)
		}
		reg := &consulapi.CatalogRegistration{
			Node:    s.Namespace,
			Address: s.Spec.ClusterIP,
			Service: service,
			// Check: &consulapi.AgentCheck{
			// 	ServiceName: s.GetName(),
			// 	Name:        s.GetName() + " health check.",
			// 	Status:      "unknown",
			// },
		}
		wm, err := kc.consulCatalog.Register(reg, &consulapi.WriteOptions{})
		if err != nil {
			log.Println("Error registering service:", err)
		} else {
			log.Println(wm)
		}
	}
}

// removeService deregisters a kubernetes service in Consul
func (kc *kube2consul) removeService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		log.Printf("Remove Service %+v\n", s.GetName())
		service := &consulapi.CatalogDeregistration{
			ServiceID: s.GetName(),
		}
		_, err := kc.consulCatalog.Deregister(service, &consulapi.WriteOptions{})
		if err != nil {
			log.Println("Error registering service:", err)
		}
	}
}

func (kc *kube2consul) updateService(oldObj, obj interface{}) {
	kc.removeService(oldObj)
	kc.newService(obj)
}
