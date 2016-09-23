package interfaces

import ()

type Kube2Consul interface {
	NodeIPByPodIP(string) string
}
