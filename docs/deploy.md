
## Deployment

```
# deploy fastcfs driver
$ deploy/kubernetes-1.20/deploy.sh
```

## Prerequisites

* Enable FastCFS Auth Service

* The FastCFS configuration file is provided by url

## Run example application and validate

Next, validate the deployment.  First, ensure all expected pods are running properly including
the external attacher, provisioner, resizer and the actual FastCFS driver plugin:

```shell
$ kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
csi-fcfs-attacher-0      1/1     Running   30         7d5h
csi-fcfs-provisioner-0   1/1     Running   3          66m
csi-fcfs-resizer-0       1/1     Running   15         3h50m
csi-fcfs-socat-0         1/1     Running   0          22m
csi-fcfsplugin-0         3/3     Running   0          22m
```

From the root directory, deploy the application pods including a config map, a secret, a storage class, a PVC, and a pod
which mounts a volume using the FastCFS driver found in directory `./examples`:

### Creating CSI configuration

1. update `csi-config-map-sample.yaml`

2. update `secret.yaml`

```shell
kubectl apply -f examples/csi-config-map-sample.yaml
kubectl apply -f examples/csi-secret.yaml
```

### Run example application and validate

```shell
kubectl apply -f examples/csi-storageclass.yaml
kubectl apply -f examples/csi-pvc.yaml
kubectl apply -f examples/csi-app.yaml
```

Let's validate the components are deployed:
```shell
$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-007e7d7e-4586-4c52-a7b5-a54694aee194   1Gi        RWX            Delete           Bound    default/csi-fcfs-pvc   csi-fcfs-sc             13s
```

```shell
$ kubectl get pvc
NAME           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
csi-fcfs-pvc   Bound    pvc-007e7d7e-4586-4c52-a7b5-a54694aee194   1Gi        RWX            csi-fcfs-sc    9s
```

Finally, inspect the application pod `my-csi-app`  which mounts a FastCFS volume:

```shell
$ kubectl describe po my-csi-app
Name:         my-csi-app
Namespace:    default
Priority:     0
Node:         kind-control-plane/172.18.0.2
Start Time:   Sun, 30 May 2021 22:47:21 +0800
Labels:       <none>
Annotations:  <none>
Status:       Running
IP:           10.244.0.18
IPs:
  IP:  10.244.0.18
Containers:
  my-frontend:
    Container ID:  containerd://74bf6c52cf48485aa781ef418072c387fe978e771dbc822e3e9d6b2fa5df2f9e
    Image:         busybox
    Image ID:      docker.io/library/busybox@sha256:b5fc1d7b2e4ea86a06b0cf88de915a2c43a99a00b6b3c0af731e5f4c07ae8eff
    Port:          <none>
    Host Port:     <none>
    Command:
      sleep
      1000000
    State:          Running
      Started:      Sun, 30 May 2021 22:47:41 +0800
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /data from my-csi-volume (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-8ld8w (ro)
Conditions:
  Type              Status
  Initialized       True 
  Ready             True 
  ContainersReady   True 
  PodScheduled      True 
Volumes:
  my-csi-volume:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  csi-fcfs-pvc
    ReadOnly:   false
  default-token-8ld8w:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-8ld8w
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                 node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason                  Age   From                     Message
  ----    ------                  ----  ----                     -------
  Normal  Scheduled               33s   default-scheduler        Successfully assigned default/my-csi-app to kind-control-plane
  Normal  SuccessfulAttachVolume  33s   attachdetach-controller  AttachVolume.Attach succeeded for volume "pvc-007e7d7e-4586-4c52-a7b5-a54694aee194"
  Normal  Pulling                 17s   kubelet                  Pulling image "busybox"
  Normal  Pulled                  14s   kubelet                  Successfully pulled image "busybox" in 3.078241754s
  Normal  Created                 13s   kubelet                  Created container my-frontend
  Normal  Started                 13s   kubelet                  Started container my-frontend
```

## Confirm FastCFS driver works

A file written in a properly mounted fastcfs volume inside an application should show up inside the FastCFS container.
The following steps confirms that FastCFS is working properly.  
First, create a file from the application pod as shown:
```shell
$ kubectl exec -it my-csi-app -- /bin/sh
/ # touch /data/hello-world
/ # exit
```

next, ssh into the FastCFS container
```shell
$ kubectl exec -it csi-fcfsplugin-0 -c fcfs -- bash
# df -h | grep globalmount
/dev/fuse       1.0G     0  1.0G   0% /var/lib/kubelet/plugins/kubernetes.io/csi/pv/pvc-007e7d7e-4586-4c52-a7b5-a54694aee194/globalmount
# ls /var/lib/kubelet/plugins/kubernetes.io/csi/pv/pvc-007e7d7e-4586-4c52-a7b5-a54694aee194/globalmount
hello-world
```

then, start `fcfs_fused` to verify that the file shows up there:

```shell
$ fcfs_fused

Usage: fcfs_fused [options] <config_file> [start | stop | restart]

options:
	-u | --user: the username
	-k | --key: the secret key filename
	-n | --namespace: the FastDIR namespace
	-m | --mountpoint: the mountpoint
	-b | --base-path: the base path
	-N | --no-daemon: run in foreground
	-V | --version: show version info
	-h | --help: for this usage
```

```shell
sudo fcfs_fused -b ~/fcfs/bar/bp -m ~/fcfs/bar/fuse \
-k ~/admin.key -u admin \
-n csi-vol-pvc-007e7d7e-4586-4c52-a7b5-a54694aee194 \
http://192.168.99.181:8080/fastcfs/fcfs/fuse.conf restart

$ ls ~/fcfs/bar/fuse
hello-world
```

## Confirm the creation of the VolumeAttachment object
An additional way to ensure the driver is working properly is by inspecting the VolumeAttachment API object created
that represents the attached volume:

```shell
$ kubectl describe volumeattachment
Name:         csi-9808ccbff0ebcc724c0ad7fe929a71b43dfc942ce75187bb10c637bb2bb4d652
Namespace:    
Labels:       <none>
Annotations:  <none>
API Version:  storage.k8s.io/v1
Kind:         VolumeAttachment
Metadata:
  Creation Timestamp:  2021-05-30T15:09:42Z
  Managed Fields:
    API Version:  storage.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        f:attached:
    Manager:      csi-attacher
    Operation:    Update
    Time:         2021-05-30T15:09:42Z
    API Version:  storage.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:attacher:
        f:nodeName:
        f:source:
          f:persistentVolumeName:
    Manager:         kube-controller-manager
    Operation:       Update
    Time:            2021-05-30T15:09:42Z
  Resource Version:  264780
  UID:               25f3adaf-84ad-474d-9c52-ef445c2d3c75
Spec:
  Attacher:   fcfs.csi.vazmin.github.io
  Node Name:  kind-control-plane
  Source:
    Persistent Volume Name:  pvc-007e7d7e-4586-4c52-a7b5-a54694aee194
Status:
  Attached:  true
Events:      <none>
```


Have Fun!