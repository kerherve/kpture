package kubernetes

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AlecAivazis/survey/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func SelectPod(kubeclient *kubernetes.Clientset, namespace string, kconfig *rest.Config) ([]string, string) {

	pods, err := kubeclient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	listpodString := []string{}
	listpodselected := []string{}

	for _, pod := range pods.Items {
		listpodString = append(listpodString, pod.Name)
	}

	prompt := &survey.MultiSelect{Options: listpodString, PageSize: 30}
	survey.AskOne(prompt, &listpodselected, survey.WithPageSize(10))

	service, err := kubeclient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{LabelSelector: "service=kpture-proxy-service"})
	if err != nil {
		fmt.Println(err)
	}

	u, err := url.Parse(kconfig.Host)
	if err != nil {
		panic(err)
	}
	return listpodselected, u.Hostname() + ":" + fmt.Sprint(service.Items[0].Spec.Ports[0].NodePort)
}
