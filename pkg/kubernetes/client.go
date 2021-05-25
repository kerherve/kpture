package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//LoadClient Load kubernetes client from kubeconfig
func LoadClient(path string) (*kubernetes.Clientset, error) {
	kconfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	// create the clientset
	kubeclient, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return nil, err
	}
	return kubeclient, err
}

//LoadConfig Load kubernetes configuration
func LoadConfig(path string) (*rest.Config, error) {
	kconfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	return kconfig, err
}
