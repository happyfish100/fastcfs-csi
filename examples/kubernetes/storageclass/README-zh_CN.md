# 配置 StorageClass

[English](./README.md) | 简体中文

此示例演示如何配置 Kubernetes 存储类以使用各种配置参数供应 FastCFS 卷。

## 使用
1. 编辑 [example manifest](./specs/example.yaml)  中的 StorageClass 配置并将 storageclass 参数更新为所需的值。

2. 发布示例:
```sh
kubectl apply -f specs/
```

3. 验证卷已创建:
```sh
kubectl describe pv
```

4. 验证 pod 成功将数据写入卷:
```sh
kubectl exec -it app cat /data/out.txt
```

5. 清理资源:
```sh
kubectl delete -f specs/
```
