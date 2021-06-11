# Static Provisioning 
This example shows how to create and consume persistent volume from exising FastCFS using static provisioning.

## Usage
1. Edit the PersistentVolume spec in [example manifest](./specs/example.yaml). Update `volumeHandle` with FastCFS pool name that you are going to use. 

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: test-pv
spec:
  capacity:
    storage: 4Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  storageClassName: ""
  csi:
    driver: fcfs.csi.vazmin.github.io
    fsType: ext4
    volumeHandle: {volumeId}
    nodeStageSecretRef:
      name: csi-fcfs-secret
      namespace: default
    volumeAttributes:
      "clusterID": virtual-cluster-id-1
      "static": "true"
    persistentVolumeReclaimPolicy: Retain
```

2. Deploy the example:
```sh
kubectl apply -f specs/
```

3. Verify application pod is running:
```sh
kubectl describe po app
```

4. Validate the pod successfully wrote data to the volume:
```sh
kubectl exec -it app cat /data/out.txt
```

5. Cleanup resources:
```sh
kubectl delete -f specs/
```
