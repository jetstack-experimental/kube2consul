package interfaces

import (
	kapi "k8s.io/kubernetes/pkg/api"
)

type Kube2Consul interface {
	NodeByPodIP() *kapi.Node
}
