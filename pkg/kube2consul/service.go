package kube2consul

import (
	"reflect"

	log "github.com/Sirupsen/logrus"
	kapi "k8s.io/kubernetes/pkg/api"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
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
	go serviceController.Run(k.stopCh)
	return serviceStore
}

func (k *Kube2Consul) newService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		log.Debugf("add service %s/%s", s.Namespace, s.Name)
		k.registerService(s)
	}
}

func (k *Kube2Consul) removeService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		log.Debugf("remove service %s/%s", s.Namespace, s.Name)
		k.registerService(nil)
	}
}

func (k *Kube2Consul) updateService(oldObj, obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok && !reflect.DeepEqual(oldObj, obj) {
		log.Debugf("update service %s/%s", s.Namespace, s.Name)
		k.registerService(s)
	}
}

func (k *Kube2Consul) registerService(kservice *kapi.Service) {
	svc := k.getOrCreateService(
		kservice.Namespace,
		kservice.Name,
	)
	svc.UpdateService(kservice)
	svc.Update()
}
