## Volume Resizing
This example shows how to resize FastCFS persistence volume using volume resizing features.

## Usage
1. Add `allowVolumeExpansion: true` in the StorageClass spec in [example manifest](./spec/example.yaml) to enable volume expansion. You can only expand a PVC if its storage class’s allowVolumeExpansion field is set to true

2. Deploy the example:
```sh
kubectl apply -f specs/
``` 

3. Verify the volume is created and Pod is running:
```sh
kubectl get pv
kubectl get po app
```

4. Expand the volume size by increasing the capacity in PVC's `spec.resources.requests.storage`:
```sh
kubectl edit pvc fcfs-claim
```
Save the result at the end of the edit.

5. Verify that both the persistence volume and persistence volume claim are resized:
```sh
kubectl get pv
kubectl get pvc
```
You should see that both should have the new value relfected in the capacity fields.

6. Verify that the application is continuously running without any interruption:
```sh
kubectl exec -it app cat /data/out.txt
```

7. Cleanup resources:
```
kubectl delete -f specs/
```
