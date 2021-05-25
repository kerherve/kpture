package kubernetes

import (
	"context"
	"fmt"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetNodeProxyNodePort(kubeclient *kubernetes.Clientset, kconfig *rest.Config) (string, error) {
	nodeport, err := getNodeProxyNodePort(kubeclient)
	if err != nil {
		return "", err
	}
	ip, err := getMasterIP(kconfig)
	if err != nil {
		return "", err
	}

	return ip + ":" + nodeport, nil
}

//GetNodeProxyNodePort return the kube proxy nodeport
func getNodeProxyNodePort(kubeclient *kubernetes.Clientset) (string, error) {
	service, err := kubeclient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{LabelSelector: "service=kpture-proxy-service"})
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return fmt.Sprint(service.Items[0].Spec.Ports[0].NodePort), nil
}

//GetMasterIP return the master ip from the kubeconfig
func getMasterIP(kconfig *rest.Config) (string, error) {
	u, err := url.Parse(kconfig.Host)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
