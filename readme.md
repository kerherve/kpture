# Kpture, packet capture for k8s

<img src="./logo/logo.png" width="100">

----

Kpture is a set of software built to help capturing packet in Kubernetes environments. 


## Get Kpture

```go
go get github.com/kpture/kpture
```

## Install Kpture in your Kubernetes cluster

This will install a daemonset handling the packet capture on the node where the targets pod are living as well as proxy pod to reach the correct daemonset pod depending on the request.

You will need to know your containerd socket location as well as the containerd namespace the pod are living

```go
$ ctr ns list
NAME   LABELS
k8s.io

$ kpture install
? Containerd socket location: /run/containerd/containerd.sock
? Containerd namespace: k8s.io
```

Check your installation

```
$ kubectl get pods -n kpture
NAME                            READY   STATUS    RESTARTS   AGE
kpture-ds-ztgxp                 1/1     Running   1          9h
kpture-proxy-7bb7f5c494-4hzxw   1/1     Running   1          9h
```

## Start a capture

```
$ kpture -o out
? [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [x]  nging2-xc8zt
  [x]  nging-87ssj
```



Stop the capture by stopping the process Ctrl^c, each pcap file will be located on the output folder

```
$ ls out
nging-87ssj.pcap  nging2-xc8zt.pcap
```

