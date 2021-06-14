## 卷大小调整

[English](./README.md) | 简体中文

此示例展示如何使用卷大小调整功能来调整 FastCFS 持久卷的大小。

## 使用
1. 在 [example manifest](./spec/example.yaml) 的 StorageClass 规范中添加 `allowVolumeExpansion: true` 以启用卷扩展。 如果存储类的 allowVolumeExpansion 字段设置为 true，则可以扩展 PVC

2. 发布示例:
```sh
kubectl apply -f specs/
``` 

3. 验证卷已创建并且 Pod 正在运行:
```sh
kubectl get pv
kubectl get po app
```

4. 通过增加 PVC 的`spec.resources.requests.storage` 中的容量来扩展卷大小:
```sh
kubectl edit pvc fcfs-claim
```
在编辑结束时保存结果。

5. 验证持久卷和持久卷声明都已调整大小:
```sh
kubectl get pv
kubectl get pvc
```
您应该会看到两者都应在容量字段中反映新值。

6. 验证应用程序是否连续运行而没有任何中断:
```sh
kubectl exec -it app cat /data/out.txt
```

7. 清理资源:
```
kubectl delete -f specs/
```
