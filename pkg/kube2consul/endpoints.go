package kube2consul

import (
	"reflect"

	log "github.com/Sirupsen/logrus"
	kapi "k8s.io/kubernetes/pkg/api"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
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
	go endpointsController.Run(k.stopCh)
	return endpointsStore
}

func (k *Kube2Consul) newEndpoints(obj interface{}) {
	if s, ok := obj.(*kapi.Endpoints); ok {
		log.Debugf("add endpoints %s/%s", s.Namespace, s.Name)
		k.registerEndpoints(s)
	}
}

func (k *Kube2Consul) removeEndpoints(obj interface{}) {
	if s, ok := obj.(*kapi.Endpoints); ok {
		log.Debugf("remove endpoints %s/%s", s.Namespace, s.Name)
		k.registerEndpoints(nil)
	}
}

func (k *Kube2Consul) updateEndpoints(oldObj, obj interface{}) {
	if s, ok := obj.(*kapi.Endpoints); ok && !reflect.DeepEqual(oldObj, obj) {
		log.Debugf("update endpoints %s/%s", s.Namespace, s.Name)
		k.registerEndpoints(s)
	}
}

func (k *Kube2Consul) registerEndpoints(kendpoints *kapi.Endpoints) {
	svc := k.getOrCreateService(
		kendpoints.Namespace,
		kendpoints.Name,
	)
	svc.UpdateEndpoints(kendpoints)
	svc.Update()
}
