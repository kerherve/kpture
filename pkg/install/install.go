package install

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

//InstallDaemonset Create the kubernetes daemonset
func InstallDaemonset(Client *kubernetes.Clientset, ns string) error {
	_, errns := Client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, metav1.CreateOptions{})

	if errns != nil {
		fmt.Println(errns)
	}

	p := true
	tm := metav1.TypeMeta{APIVersion: "apps/v1"}
	om := metav1.ObjectMeta{Name: "kpture-ds", Namespace: ns, Labels: map[string]string{"name": "kptureDs"}}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"name": "kptureDs"}}
	container := []corev1.Container{corev1.Container{
		Name: "kpture-ds", Image: "gmtstephane/kpture-server:v0.2.0",
		Ports: []corev1.ContainerPort{corev1.ContainerPort{ContainerPort: 8080}},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{Name: "ctrsock", MountPath: "/var/snap/microk8s/common/run/containerd.sock"},
			corev1.VolumeMount{Name: "proc", MountPath: "/proc/"},
		},
		SecurityContext: &corev1.SecurityContext{Privileged: &p},
	}}
	volumes := []corev1.Volume{
		corev1.Volume{Name: "ctrsock", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/var/snap/microk8s/common/run/containerd.sock"}}},
		corev1.Volume{Name: "proc", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/proc/"}}},
	}

	podSpec := corev1.PodSpec{Volumes: volumes, Containers: container}
	podTemplate := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"name": "kptureDs"}}, Spec: podSpec}

	ds := v1.DaemonSet{TypeMeta: tm, ObjectMeta: om, Spec: v1.DaemonSetSpec{Selector: &labelSelector, Template: podTemplate}}

	_, err := Client.AppsV1().DaemonSets("kpture").Create(context.Background(), &ds, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	}
	if errns != nil {
		return errors.Wrap(err, errns.Error())
	}
	return err
}

func InstallProxy(Client *kubernetes.Clientset, ns string) error {
	tm := metav1.TypeMeta{APIVersion: "apps/v1"}
	om := metav1.ObjectMeta{Name: "kpture-proxy"}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"app": "kpture-proxy"}}
	container := []corev1.Container{corev1.Container{
		Name: "kpture-proxy", Image: "gmtstephane/kpture-proxy:v0.2.0",
		Ports: []corev1.ContainerPort{corev1.ContainerPort{ContainerPort: 8080}},
	}}
	podSpec := corev1.PodSpec{Containers: container}
	podTemplate := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "kpture-proxy"}}, Spec: podSpec}

	ds := v1.Deployment{TypeMeta: tm, ObjectMeta: om, Spec: v1.DeploymentSpec{Selector: &labelSelector, Template: podTemplate}}

	_, err := Client.AppsV1().Deployments(ns).Create(context.Background(), &ds, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Installservice(Client *kubernetes.Clientset, ns string) error {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "kpture-proxy-service", Labels: map[string]string{"service": "kpture-proxy-service"}}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "kpture-proxy"}, Type: corev1.ServiceTypeNodePort, Ports: []corev1.ServicePort{corev1.ServicePort{Port: 8080, TargetPort: intstr.IntOrString{IntVal: 8080}}}}}
	_, err := Client.CoreV1().Services(ns).Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
