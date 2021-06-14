# Dynamic Volume Provisioning
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

3. Validate the pod successfully wrote data to the volume:
```
kubectl exec -it app cat /data/out.txt
```

4. Cleanup resources:
```
kubectl delete -f specs/pod.yaml
kubectl delete -f specs/claim.yaml
kubectl delete -f specs/storageclass.yaml
```
