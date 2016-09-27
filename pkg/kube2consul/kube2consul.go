package kube2consul

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	kapi "k8s.io/kubernetes/pkg/api"
	kubernetes "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	krest "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"

	"github.com/simonswine/kube2consul/pkg/detect_node"
	"github.com/simonswine/kube2consul/pkg/interfaces"
	"github.com/simonswine/kube2consul/pkg/service"
)

var AppVersion string = "unknown"
var AppName string = "kube2consul"

type Kube2Consul struct {
	RootCmd             *cobra.Command
	kubernetesClient    *kclient.Client
	kubernetesClientset *kubernetes.Clientset
	kubernetesConfig    *krest.Config
	Kubeconfig          string
	consulClient        *consulapi.Client
	consulCatalog       *consulapi.Catalog
	consulAddress       string
	detectNode          *detect_node.DetectNode
	resyncPeriod        time.Duration

	services     map[string]*service.Service
	servicesLock sync.Mutex

	// stop channel for shutting down
	stopCh chan struct{}

	// wait group
	waitGroup sync.WaitGroup
}

var _ interfaces.Kube2Consul = &Kube2Consul{}

func New() *Kube2Consul {
	k := &Kube2Consul{
		resyncPeriod: 5 * time.Minute,
		stopCh:       make(chan struct{}),
		waitGroup:    sync.WaitGroup{},
		services:     make(map[string]*service.Service),
	}
	k.init()
	return k
}

func (k *Kube2Consul) userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func (k *Kube2Consul) NodeIPByPodIP(podIP string) (nodeIP string, err error) {
	return k.detectNode.NodeIPByPodIP(podIP)
}

func (k *Kube2Consul) NodeNameByPodIP(podIP string) (nodeName string, err error) {
	return k.detectNode.NodeNameByPodIP(podIP)
}

func (k *Kube2Consul) init() {

	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	k.RootCmd = &cobra.Command{
		Use:   "kube2consul",
		Short: "Export Kubernetes NodePort services to Consul",
		Run: func(cmd *cobra.Command, args []string) {
			k.cmdRun()
		},
	}
	k.RootCmd.PersistentFlags().StringVarP(
		&k.Kubeconfig,
		"kubeconfig",
		"k",
		filepath.Join(k.userHomeDir(), ".kube/config"),
		"path to the kubeconfig file",
	)

	k.RootCmd.PersistentFlags().StringVarP(
		&k.consulAddress,
		"consoul-address",
		"c",
		"localhost:8500",
		"consoul server address",
	)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print the version number of %s", AppVersion),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", AppName, AppVersion)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List servicecs that whould have been registered in consul",
		Run: func(cmd *cobra.Command, args []string) {
			k.cmdList()
		},
	}

	k.RootCmd.AddCommand(versionCmd)
	k.RootCmd.AddCommand(listCmd)

	k.detectNode = detect_node.New(k)

}

func (k *Kube2Consul) cmdList() {
	svcs, err := k.KubernetesClient().Services(kapi.NamespaceAll).List(kapi.ListOptions{})

	if err != nil {
		log.Warnf("Error getting services: %s", err)
	}

	for _, svc := range svcs.Items {
		if kapi.ServiceType(svc.Spec.Type) == kapi.ServiceTypeNodePort {
			endpoints, err := k.KubernetesClient().Endpoints(svc.Namespace).Get(svc.Name)
			if err != nil {
				log.Warnf(
					"Error getting endpoints for %s/%s: %s",
					svc.Namespace,
					svc.Name,
					err,
				)
				continue
			}

			s := service.New(k, svc.Namespace, svc.Name)
			s.UpdateService(&svc)
			s.UpdateEndpoints(endpoints)
			list := s.List()
			for _, elem := range list {
				fmt.Printf("%+v\n", elem)
			}

		}
	}
}

func (k *Kube2Consul) cmdRun() {
	k.watchForServices()
	k.watchForEndpointss()
	select {}
}

func (k *Kube2Consul) KubernetesConfig() *krest.Config {
	if k.kubernetesConfig == nil {
		// try in cluster first
		config, err := krest.InClusterConfig()
		if err != nil {
			log.Warnf("Failed to connect to kubernetes using in-cluster configuration: %s", err)
			log.Infof("Trying configuration from %s", k.Kubeconfig)
			// uses the current context in kubeconfig
			config, err = kclientcmd.BuildConfigFromFlags("", k.Kubeconfig)
			if err != nil {
				panic(err.Error())
			}
		}
		k.kubernetesConfig = config
	}
	return k.kubernetesConfig
}

func (k *Kube2Consul) KubernetesClient() *kclient.Client {
	if k.kubernetesClient == nil {
		client, err := kclient.New(k.KubernetesConfig())
		if err != nil {
			panic(err.Error())
		}
		k.kubernetesClient = client
	}
	return k.kubernetesClient
}

func (k *Kube2Consul) KubernetesClientset() *kubernetes.Clientset {
	if k.kubernetesClientset == nil {
		clientset, err := kubernetes.NewForConfig(k.KubernetesConfig())
		if err != nil {
			panic(err.Error())
		}
		k.kubernetesClientset = clientset
	}
	return k.kubernetesClientset
}

func (k *Kube2Consul) getOrCreateService(namespace string, name string) *service.Service {
	key := fmt.Sprintf("%s/%s", namespace, name)

	k.servicesLock.Lock()
	defer k.servicesLock.Unlock()
	if svc, ok := k.services[key]; ok {
		return svc
	}
	svc := service.New(
		k,
		namespace,
		name,
	)
	k.services[key] = svc
	return svc
}
