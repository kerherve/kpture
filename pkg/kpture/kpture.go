package kpture

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/kpture/kpture/pkg/kubernetes"
	"github.com/kpture/kpture/pkg/socket"
	"github.com/kpture/kpture/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

type KptureCli struct {
	Logger        *log.Entry
	K8sclient     *kubernetes.KubernetesClient
	SelectedPods  *[]v1.Pod
	CaptureStop   chan bool
	GlobalCapture chan []byte
	outputFolder  string
	configPath    string
	namespaces    []string
	loglevel      log.Level
	socket        *socket.SocketCapture
}

func (k *KptureCli) Setup(configPath string, outputFolder string, namespaces []string, loglevel log.Level, global chan []byte) {
	k.outputFolder = outputFolder
	k.configPath = configPath
	k.namespaces = namespaces
	k.loglevel = loglevel
	k.CaptureStop = make(chan bool)
	k.GlobalCapture = global

	k.Logger = utils.NewLogger("cli", loglevel)
	k.generateK8sClient()
	k.getSelectedPods()
	k.setupFolders()
	k.HandleLogCapture()
	k.setupSignalHandler()

}

func (k *KptureCli) getKptureProxyAddress() string {
	service, err := k.K8sclient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{LabelSelector: "service=kpture-proxy-service"})
	if err != nil {
		fmt.Println(err)
	}

	u, err := url.Parse(k.K8sclient.Host)
	if err != nil {
		panic(err)
	}

	k.Logger.Info("Kpture Proxy is reachable at " + u.Hostname() + ":" + fmt.Sprint(service.Items[0].Spec.Ports[0].NodePort))

	return u.Hostname() + ":" + fmt.Sprint(service.Items[0].Spec.Ports[0].NodePort)
}
func (k *KptureCli) Start(metricserver bool) {
	proxyAddr := k.getKptureProxyAddress()

	k.socket = socket.NewSocketCapture(k.loglevel)
	for _, pod := range *k.SelectedPods {
		k.socket.AddCapture(&socket.Capture{CaptureInfo: &socket.CaptureInfo{ContainerName: pod.Name, ContainerNamespace: pod.Namespace, Interface: "eth0", FileName: k.outputFolder + "/" + pod.Name + "/" + pod.Name + ".pcap"}})
	}
	k.socket.StartCapture(proxyAddr, k.GlobalCapture)
	if metricserver {
		go k.socket.MetricServer()
	}
	for {

	}
}

func (k *KptureCli) generateK8sClient() {
	var err error
	k.K8sclient, err = kubernetes.NewKubernetesClientFromConfig(k.configPath, k.Logger.Level)
	if err != nil {
		log.Fatalf("Error loading kubernetes client", err)
	}
}

func (k *KptureCli) getSelectedPods() {
	k.SelectedPods = k.K8sclient.SelectPod(k.namespaces)
	if len(*k.SelectedPods) == 0 {
		k.Logger.Info("No pod found")
		os.Exit(1)
		return
	}
}

func (k *KptureCli) HandleLogCapture() {
	podLogOpts := v1.PodLogOptions{}
	podLogOpts = v1.PodLogOptions{SinceTime: &metav1.Time{Time: time.Now()}}
	go func() {
		<-k.CaptureStop
		k.Logger.Info("Saving logs")
		k.K8sclient.GetLogs(k.SelectedPods, podLogOpts, k.outputFolder)
		os.Exit(0)
	}()
}

func (k *KptureCli) StopCapture() {
	k.CaptureStop <- true
}

func (k *KptureCli) setupSignalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		k.Logger.Info("Detected Sigterm, doing cleanup")
		k.CaptureStop <- true
	}()
}

func (k *KptureCli) setupFolders() {
	k.Logger.Trace("Creating folders " + k.outputFolder)
	err := os.Mkdir(k.outputFolder, 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			k.Logger.Info(err)
		} else {
			k.Logger.Error(err)
		}
	}
	for _, pod := range *k.SelectedPods {
		err := os.Mkdir(path.Join(k.outputFolder, pod.Name), 0755)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				k.Logger.Info(err)
			} else {
				k.Logger.Error(err)
			}
		}
	}
}

func (k *KptureCli) GetDnsHostFile() ([]byte, error) {
	var hostFile = ""
	nss, err := k.K8sclient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []byte{}, err
	}
	for _, ns := range nss.Items {
		pods, err := k.K8sclient.CoreV1().Pods(ns.GetObjectMeta().GetName()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return []byte{}, err
		}
		for _, pod := range pods.Items {
			hostFile += pod.Status.PodIP + " " + pod.GetObjectMeta().GetName() + "(" + ns.GetObjectMeta().GetName() + ")" + "\n"
		}

		svcs, err := k.K8sclient.CoreV1().Services(ns.GetObjectMeta().GetName()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return []byte{}, err
		}
		for _, svc := range svcs.Items {
			hostFile += svc.Spec.ClusterIP + " " + svc.GetName() + "(" + ns.GetObjectMeta().GetName() + ")" + "\n"
		}
	}

	return []byte(hostFile), nil
}
