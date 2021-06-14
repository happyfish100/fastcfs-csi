# Dynamic Volume Provisioning

English | [简体中文](./README-zh_CN.md)

This example shows how to create a FastCFS volume and consume it from container dynamically.

## Usage

1. Create a sample app along with the StorageClass and the PersistentVolumeClaim:
```
kubectl apply -f specs/
```

2. Validate the volume was created and `volumeHandle` contains an FastCFS volumeID:
```
kubectl describe pv
```
e.g, volumeID is `0014-virtual-cluster-id-1-0005-admin-0030-csi-vol-pvc-7c6b596e-e158-43e8-8553-d23d39ed46f8` (clusterID length + `clusterID` + userName length + `userName` + poolName length + `poolName`), so FastCFS `userName` is `admin`, FastCFS `poolName` is `csi-vol-pvc-7c6b596e-e158-43e8-8553-d23d39ed46f8`

3. Validate the pod successfully wrote data to the volume:
```
kubectl exec -it app cat /data/out.txt
```

  * Validate the FastCFS pool has been created on the FastCFS server
    ```
    fcfs_pool plist <userName> <poolName>
    ```

  * Validate the pod successfully wrote data to the mountpoint through the FastCFS client
    ```
    mkdir ~/foo
    fcfs_fused -n <poolName> -m ~/foo /etc/fastcfs/fcfs/fuse.conf restart
    cat ~/foo/out.txt
    ```

4. Cleanup resources:
```
kubectl delete -f specs/pod.yaml
kubectl delete -f specs/claim.yaml
kubectl delete -f specs/storageclass.yaml
```
