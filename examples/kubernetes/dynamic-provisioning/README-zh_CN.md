# 动态卷供应

[English](./README.md) | 简体中文

此示例演示如何创建 FastCFS 卷并从容器动态使用它。

## 使用

1. 创建一个示例应用程序以及 StorageClass 和 PersistentVolumeClaim:
```
kubectl apply -f specs/
```

2. 验证卷是否已创建并且 `volumeHandle` 包含一个 FastCFS volumeID:
```
kubectl describe pv
```
e.g, volumeID 为 `0014-virtual-cluster-id-1-0005-admin-0030-csi-vol-pvc-7c6b596e-e158-43e8-8553-d23d39ed46f8`
(clusterID length + `clusterID` + userName length + `userName` + poolName length + `poolName`),
即 FastCFS `userName` 为 `admin`, FastCFS `poolName`为 `csi-vol-pvc-7c6b596e-e158-43e8-8553-d23d39ed46f8`

3. 验证 pod 成功将数据写入卷:
```
kubectl exec -it app cat /data/out.txt
```

  * 在 FastCFS 服务器验证 FastCFS pool 已创建
    ```
    fcfs_pool plist <userName> <poolName>
    ```

  * 通过 FastDFS 客户端验证 pod 成功写入数据到挂载点
    ```
    mkdir ~/foo
    fcfs_fused -n <poolName> -m ~/foo /etc/fastcfs/fcfs/fuse.conf restart
    cat ~/foo/out.txt
    ```

4. 清理资源:
```
kubectl delete -f specs/pod.yaml
kubectl delete -f specs/claim.yaml
kubectl delete -f specs/storageclass.yaml
```
