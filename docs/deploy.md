## Cluster setup


## Deployment


```
# deploy fastcfs driver
$ deploy/kubernetes-1.20/deploy.sh
```


## Run example application and validate

Next, validate the deployment.  First, ensure all expected pods are running properly including the external attacher, provisioner and the actual FastCFS driver plugin:

```shell
$ kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
csi-fcfs-attacher-0      1/1     Running   30         7d5h
csi-fcfs-provisioner-0   1/1     Running   3          66m
csi-fcfs-resizer-0       1/1     Running   15         3h50m
csi-fcfs-socat-0         1/1     Running   0          22m
csi-fcfsplugin-0         3/3     Running   0          22m
```

From the root directory, deploy the application pods including a storage class, a PVC, and a pod which mounts a volume using the Hostpath driver found in directory `./examples`:

### Config

1. Modify `csi-config-map-sample.yaml`

2. Modify `secret.yaml`

### Run

```shell
kubectl -f examples/csi-config-map-sample.yaml
kubectl -f examples/csi-secret.yaml
kubectl -f examples/csi-storageclass.yaml
kubectl -f examples/csi-pvc.yaml
kubectl -f examples/csi-app.yaml
```

## Confirm FastCFS driver works


## Confirm the creation of the VolumeAttachment object



