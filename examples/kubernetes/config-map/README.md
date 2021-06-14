# Config Map

English | [简体中文](./README-zh_CN.md)

This example shows how to create a config map

`clusterID` is virtualized by the user.

`configURL` is an http/https address. FastCFS client, auth module, cluster and other configuration files need to provide HTTP or HTTPS access.

The following is the directory structure and required configuration files:
```
<configURL>
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
e.g., if the `configURL` is `http://192.168.99.170:8080`, the driver will pass `http://192.168.99.170:8080/fastcfs/fcfs/fuse.conf` to the FastCFS client.

## Usage

1. Edit the ConfigMap spec in [example manifest](./specs/example.yaml). 
   Update `configURL` with FastCFS Config URL that you are going to use.
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

2. Verify the ConfigMap is created:
```sh
kubectl get cm
```

3. Cleanup resources:
```
kubectl delete -f specs/
```
