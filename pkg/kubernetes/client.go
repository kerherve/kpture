package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/kpture/kpture/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	*kubernetes.Clientset
	*rest.Config
	Logger *log.Entry
}

func NewKubernetesClientFromConfig(path string, level log.Level) (*KubernetesClient, error) {
	client := KubernetesClient{}

	client.Logger = utils.NewLogger("k8sclient", level)

	if err := client.LoadClient(path); err != nil {
		return nil, err
	}
	if err := client.LoadConfig(path); err != nil {
		return nil, err
	}
	return &client, nil
}

//LoadClient Load kubernetes client from kubeconfig
func (k *KubernetesClient) LoadClient(path string) error {
	k.Logger.WithFields(log.Fields{"configPath": path}).Trace("Loading kubernetes client from file")
	kconfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return err
	}
	// create the clientset
	kubeclient, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}
	k.Clientset = kubeclient

	return nil
}

//LoadConfig Load kubernetes configuration
func (k *KubernetesClient) LoadConfig(path string) error {
	k.Logger.WithFields(log.Fields{"configPath": path}).Trace("Loading kubernetes config from file")
	kconfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return err
	}
	k.Config = kconfig
	return nil
}

func (k *KubernetesClient) GetLogs(pods *[]v1.Pod, options v1.PodLogOptions, outputFolder string) error {
	k.Logger.Info("Gettings logs ")
	for _, pod := range *pods {

		options.Container = pod.Spec.Containers[0].Name
		req := k.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &options)
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		n, err := io.Copy(buf, podLogs)
		if n == 0 {
			return nil
		}
		if err != nil {
			fmt.Println(err)
			return err
		}

		f, err := os.Create(path.Join(outputFolder, pod.Name) + "/" + pod.Name + ".log")
		if err != nil {
			fmt.Println(err)
			cobra.CheckErr(err)
		}
		f.Write(buf.Bytes())
	}

	return nil
}
