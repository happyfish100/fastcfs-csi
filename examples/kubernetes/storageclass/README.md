# Configuring StorageClass

English | [简体中文](./README-zh_CN.md)

This example shows how to configure Kubernetes storageclass to provision FastCFS volumes with various configuration parameters.

## Usage
1. Edit the StorageClass spec in [example manifest](./specs/example.yaml) and update storageclass parameters to desired value.

2. Deploy the example:
```sh
kubectl apply -f specs/
```

3. Verify the volume is created:
```sh
kubectl describe pv
```

4. Validate the pod successfully wrote data to the volume:
```sh
kubectl exec -it app cat /data/out.txt
```

5. Cleanup resources:
```sh
kubectl delete -f specs/
```
