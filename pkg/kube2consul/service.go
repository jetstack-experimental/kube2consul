package kube2consul

import (
	log "github.com/Sirupsen/logrus"
	kapi "k8s.io/kubernetes/pkg/api"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"
)

func (k *Kube2Consul) watchForServices() kcache.Store {
	serviceStore, serviceController := kframework.NewInformer(
		kcache.NewListWatchFromClient(k.KubernetesClient(), "services", kapi.NamespaceAll, kselector.Everything()),
		&kapi.Service{},
		k.resyncPeriod,
		kframework.ResourceEventHandlerFuncs{
			AddFunc:    k.newService,
			DeleteFunc: k.removeService,
			UpdateFunc: k.updateService,
		},
	)
	go serviceController.Run(wait.NeverStop)
	return serviceStore
}

func (k *Kube2Consul) newService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		if s.Spec.Type != kapi.ServiceTypeNodePort {
			log.Debugf("skip service %s/%s as its not of type NodePort", s.Namespace, s.Name)
			return
		}
		log.Debugf("add service %s/%s", s.Namespace, s.Name)
	}
}

// removeService deregisters a kubernetes service in Consul
func (k *Kube2Consul) removeService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		log.Debugf("remove service %s/%s", s.Namespace, s.Name)
	}
}

func (k *Kube2Consul) updateService(oldObj, obj interface{}) {
	k.removeService(oldObj)
	k.newService(obj)
}
