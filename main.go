package main

import (
	"flag"
	"log"
	"time"

	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"

	consulapi "github.com/hashicorp/consul/api"
)

const (
	resyncPeriod = 5 * time.Minute
)

var (
	kubeCfgFile   = flag.String("kube-config", "", "Kubernetes Config File")
	consulAddress = flag.String("consul-address", "localhost:8500", "Consul Server address")
	config        *restclient.Config
)

func main() {
	flag.Parse()

	kubeClient, err := newKubeClient(*kubeCfgFile)
	if err != nil {
		log.Fatal(err)
	}

	consulClient, err := newConsulClient(*consulAddress)
	if err != nil {
		log.Fatal(err)
	}

	kc := newKube2Consul(kubeClient, consulClient)
	kc.watchForServices()

	select {}
}

// newKubeClient create a new Kubernetes API Client
func newKubeClient(kubeCfgFile string) (*kclient.Client, error) {
	rules := &kclientcmd.ClientConfigLoadingRules{ExplicitPath: kubeCfgFile}
	config, err := kclientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &kclientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := kclient.New(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// newConsulClient create a new Consul API Client
func newConsulClient(serverAddress string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = serverAddress
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
