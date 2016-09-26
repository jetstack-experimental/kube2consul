package detect_node

import (
	"flag"
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	kubernetes "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"

	"github.com/simonswine/kube2consul/pkg/interfaces"
)

type DetectNode struct {
	kube2consul interfaces.Kube2Consul
}

func New(k interfaces.Kube2Consul) *DetectNode {
	return &DetectNode{
		kube2consul: k,
	}
}

func (s *DetectNode) nodeNameByPodIP(podIP string) (nodeName string, err error) {
	pods, err := s.kube2consul.KubernetesClientset().Core().Pods("").List(kapi.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		if pod.Status.PodIP == podIP {
			return pod.Spec.NodeName, nil
		}
	}
	return "", fmt.Errorf("No pod found with podIP %s", podIP)
}

func (s *DetectNode) NodeIPByPodIP(podIP string) (nodeIP string, err error) {
	nodeName, err := s.nodeNameByPodIP(podIP)
	if err != nil {
		return "", err
	}

	node, err := s.kube2consul.KubernetesClientset().Core().Nodes().Get(nodeName)
	if err != nil {
		return "", err
	}

	return node.Status.Addresses[0].Address, nil

}

var (
	kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
)

type fakek2c struct {
	clientset *kubernetes.Clientset
}

func (f *fakek2c) Clientset() *kubernetes.Clientset {
	return f.clientset
}

func (f *fakek2c) NodeIPByPodIP(string) string {
	return "nop"
}
