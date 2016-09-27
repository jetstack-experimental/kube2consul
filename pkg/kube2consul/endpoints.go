package kube2consul

import (
	log "github.com/Sirupsen/logrus"
	kapi "k8s.io/kubernetes/pkg/api"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"
)

func (k *Kube2Consul) watchForEndpointss() kcache.Store {
	endpointsStore, endpointsController := kframework.NewInformer(
		kcache.NewListWatchFromClient(k.KubernetesClient(), "endpoints", kapi.NamespaceAll, kselector.Everything()),
		&kapi.Endpoints{},
		k.resyncPeriod,
		kframework.ResourceEventHandlerFuncs{
			AddFunc:    k.newEndpoints,
			DeleteFunc: k.removeEndpoints,
			UpdateFunc: k.updateEndpoints,
		},
	)
	go endpointsController.Run(wait.NeverStop)
	return endpointsStore
}

func (k *Kube2Consul) newEndpoints(obj interface{}) {
	if s, ok := obj.(*kapi.Endpoints); ok {
		log.Debugf("add endpoints %s/%s", s.Namespace, s.Name)
	}
}

// removeEndpoints deregisters a kubernetes endpoints in Consul
func (k *Kube2Consul) removeEndpoints(obj interface{}) {
	if s, ok := obj.(*kapi.Endpoints); ok {
		log.Debugf("remove endpoints %s/%s", s.Namespace, s.Name)
	}
}

func (k *Kube2Consul) updateEndpoints(oldObj, obj interface{}) {
	k.removeEndpoints(oldObj)
	k.newEndpoints(obj)
}
