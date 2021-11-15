package kubernetes

import (
	"context"

	"github.com/AlecAivazis/survey/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubernetesClient) SelectPod(namespaces []string) *[]v1.Pod {

	listpod := []v1.Pod{}
	listpodselected := []v1.Pod{}

	for _, ns := range namespaces {
		pods, err := k.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		listpod = append(listpod, pods.Items...)
	}

	listPodstring := []string{}

	for _, pod := range listpod {
		if len(namespaces) > 1 {
			listPodstring = append(listPodstring, pod.Name+" ("+pod.Namespace+")")
		} else {
			listPodstring = append(listPodstring, pod.Name)
		}
	}

	prompt := &survey.MultiSelect{Options: listPodstring, PageSize: 30}
	listpodselectedString := []string{}
	survey.AskOne(prompt, &listpodselectedString, survey.WithPageSize(10))

	// service, err := kubeclient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{LabelSelector: "service=kpture-proxy-service"})
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// u, err := url.Parse(kconfig.Host)
	// if err != nil {
	// 	panic(err)
	// }
	// return listpodselected, u.Hostname() + ":" + fmt.Sprint(service.Items[0].Spec.Ports[0].NodePort)

	for _, podFromList := range listpod {
		for _, selectedpodName := range listpodselectedString {
			if podFromList.Name == selectedpodName {
				listpodselected = append(listpodselected, podFromList)
			}
		}
	}

	return &listpodselected
}
