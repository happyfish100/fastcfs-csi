# Config Map

[English](./README.md) | 简体中文

此示例演示如何创建配置`ConfigMap`

`clusterID` 由用户自定义, 为了区分 FastCFS 集群配置文件.

`configURL` 一个网络路径URL.

```
<configURL>
    |__ fastcfs
        |
        |__ fcfs:
        |    |__ fuse.conf
        |
        |__ fdir: FastDIR server
        |    |__ cluster.conf
        |    |__ client.conf
        |
        |__ fstore: faststore server
            |__ cluster.conf
            |__ client.conf
```
例如, `configURL` 为 `http://192.168.99.170:8080`时, 驱动会通过 `http://192.168.99.170:8080/fastcfs/fcfs/fuse.conf` 获取 `fuse.conf`.

## 使用

1. 编辑 [ConfigMap示例](./specs/example.yaml) 中的 ConfigMap 信息。使用您要使用的 FastCFS 配置 URL 来更新 `configURL`。
```yaml
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    [
      {
        "clusterID": "virtual-cluster-id-1",
        "configURL": <configURL>
      }
    ]
metadata:
  name: fcfs-csi-config-example
```

