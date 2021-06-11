# Config Map
This example shows how to create a config map

`clusterID` is virtualized by the user.

`configURL` is a network path.
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
e.g., if get `fuse.conf` with `http://192.168.99.170:8080/fastcfs/fcfs/fuse.conf`. the `configURL` is `http://192.168.99.170:8080`

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

