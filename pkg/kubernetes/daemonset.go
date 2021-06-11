package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
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

func GetLogs(kubeclient *kubernetes.Clientset, namespace string, pods []string, options v1.PodLogOptions, outputFolder string) error {

	for _, podname := range pods {

		p, err := kubeclient.CoreV1().Pods(namespace).Get(context.Background(), podname, metav1.GetOptions{})
		options.Container = p.Spec.Containers[0].Name
		req := kubeclient.CoreV1().Pods(namespace).GetLogs(podname, &options)
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			fmt.Println(err)
			return err
		}
		f, err := os.Create(outputFolder + "/" + podname + "/" + podname + ".log")
		if err != nil {
			fmt.Println(err)
			cobra.CheckErr(err)
		}
		f.Write(buf.Bytes())
	}

	return nil
}
