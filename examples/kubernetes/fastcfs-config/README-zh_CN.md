# FastCFS Config

[English](./README.md) | 简体中文

此示例演示如何创建适用于 FastCFS CSI 的 FastCFS 配置

## 使用

FastCFS CSI可以使用以下两种方式

1. 直接使用容器的文件系统地址

定义[ConfigMap](../../../deploy/kubernetes/base/fastcfs-client-configmap.yaml), 再mount到容器即可.

详细请查看[controller.yaml](../../../deploy/kubernetes/base/controller.yaml)
或者[node.yaml](../../../deploy/kubernetes/base/node.yaml)

以下为部分配置, 仅作说明:
```yaml
containers:
- volumeMounts:
    - mountPath: /etc/fastcfs-client-config
      name: fcfs-config
volumes:
- name: fcfs-config
  configMap:
    name: fastcfs-client-config
    items:
      - key: fdir-cluster
        path: fastcfs/fdir/cluster.conf
      - key: fstore-cluster
        path: fastcfs/fstore/cluster.conf
      - key: auth-cluster
        path: fastcfs/auth/cluster.conf
      - key: auth-config
        path: fastcfs/auth/auth.conf
      - key: auth-client
        path: fastcfs/auth/client.conf
      - key: fuse-config
        path: fastcfs/fcfs/fuse.conf
```

2. 使用 http/https 路径

将FastCFS的配置文件放置到web服务器, 并将存储类的参数`fastcfs-config-base-path`设置为web服务器的访问地址

例如:
```yaml
# storageclass.yaml 其余配置省略
parameters:
  fastcfs-config-base-path: http://192.168.99.170:8080
```


## 目录结构及所需配置文件
```
<fastcfs-config-base-path>
    |
    |__ fastcfs
        |
        |__ auth:
        |    |__ auth.conf    
        |    |__ client.conf
        |    |__ cluster.conf
        |
        |__ fcfs:
        |    |__ fuse.conf
        |
        |__ fdir:
        |    |__ cluster.conf
        |
        |__ fstore:
             |__ cluster.conf
```

即 `fuse.conf` 位于 `/mypath/fastcfs/fcfs/fuse.conf` 或者 `http://ip:port/fastcfs/fcfs/fuse.conf`

注意: `fastcfs-config-base-path`的值不宜过长, 该存储类的 `fastcfs-config-base-path` 的长度和 该存储类使用的`secret`的用户名的长度总和不应超过`63`个字符.
