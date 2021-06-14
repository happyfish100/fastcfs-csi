# 静态供应

[English](./README.md) | 简体中文

此示例演示如何使用静态配置从现有 FastCFS 创建和使用持久卷。

## 使用
1. 编辑 [example manifest](./specs/example.yaml) 中的 PersistentVolume 配置。 使用您要使用的 FastCFS poolName更新`volumeHandle`。

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: test-pv
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  storageClassName: ""
  csi:
    driver: fcfs.csi.vazmin.github.io
    volumeHandle: {volumeId}
    nodeStageSecretRef:
      name: csi-fcfs-secret
      namespace: default
    volumeAttributes:
      "clusterID": virtual-cluster-id-1
      "static": "true"
  persistentVolumeReclaimPolicy: Retain
```

2. 发布示例:
```sh
kubectl apply -f specs/
```

3. 验证应用程序 pod 正在运行:
```sh
kubectl describe po app
```

4. 验证 pod 成功将数据写入卷:
```sh
kubectl exec -it app cat /data/out.txt
```

5. 清理资源:
```sh
kubectl delete -f specs/
```
