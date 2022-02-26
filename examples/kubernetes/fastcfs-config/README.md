# FastCFS Config

English | [简体中文](./README-zh_CN.md)

This example shows how to create a FastCFS configuration for FastCFS CSI

## Usage

FastCFS CSI can use the following two ways

1. Use the container's filesystem path directly

Define [ConfigMap](../../../deploy/kubernetes/base/fastcfs-client-configmap.yaml), and then mount it to the container.

More details, please see [controller.yaml](../../../deploy/kubernetes/base/controller.yaml)
or [node.yaml](../../../deploy/kubernetes/base/node.yaml)

The following is a partial configuration, just for illustration:
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

2. Use http/https paths

Place the FastCFS configuration file on the web server, and set the storage class parameter `fastcfs-config-base-path` to the access address of the web server

i.e.
```yaml
# storageclass.yaml, the rest of the configuration is omitted
parameters:
  fastcfs-config-base-path: http://192.168.99.170:8080
```


## the directory structure and required configuration files

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

i.e. `fuse.conf` is located at `/mypath//fastcfs/fcfs/fuse.conf` 
or `http://ip:port/fastcfs/fcfs/fuse.conf`

Note: The value of `fastcfs-config-base-path` should not be too long, the sum of the length of `fastcfs-config-base-path` of this storage class and the length of the username of `secret` of this storage class should not exceed `63` characters.
