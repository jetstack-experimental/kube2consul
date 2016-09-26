package interfaces

import (
	kubernetes "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

type Kube2Consul interface {
	KubernetesClientset() *kubernetes.Clientset
	KubernetesClient() *kclient.Client
	NodeIPByPodIP(string) (string, error)
}
