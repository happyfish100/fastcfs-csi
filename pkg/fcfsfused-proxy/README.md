# fcfsfused-proxy
 - supported CSI driver version: v1.1.0 or later version

By default, restart csi-fcfsfused-node daemonset would make current fcfs_fused mounts unavailable.
When fuse nodeserver restarts on the node, the fuse daemon also restarts, this results in breaking all connections FUSE daemon is maintaining.
You could find more details here: [No easy way how to update CSI driver that uses fuse](https://github.com/kubernetes/kubernetes/issues/70013).

This page shows how to run a fcfs_fused proxy on all agent nodes and this proxy mounts volumes, maintains FUSE connections. 
> fcfsfused proxy receives mount request in a GRPC call and then uses this data to mount and returns the output of the fcfs_fused command.

### Step#1. Install fcfsfused-proxy on debian based agent node
> below daemonset would also install latest [fcfs_fused](https://github.com/happyfish100/FastCFS) version on the node
```console
kubectl apply -f https://raw.githubusercontent.com/happyfish100/fastcfs-csi/master/deploy/fcfsfused-proxy/fcfsfused-proxy.yaml
```

### Step#2. Install FastFCS CSI driver with `node.enableBlobfuseProxy=true` setting

```console

```

#### Troubleshooting
 - Get `fcfsfused-proxy` logs on the node
```console
kubectl get po -n kube-system -o wide | grep fcfsfused-proxy
csi-fcfsfused-proxy-47kpp                    1/1     Running   0          37m
kubectl logs -n kube-system csi-fcfsfused-proxy-47kpp
```

#### Development
 - install fcfsfused-proxy package, run as a service manually
```console
wget https://github.com/happyfish100/fastcfs-csi/raw/master/deploy/fcfsfused-proxy/v0.1.0/fcfsfused-proxy-v0.1.0.rpm -O /tmp/fcfsfused-proxy-v0.1.0.rpm
rpm -ivh /tmp/fcfsfused-proxy-v0.1.0.rpm
mkdir -p /var/lib/kubelet/plugins/fcfs.csi.vazmin.github.io
systemctl enable fcfsfused-proxy
systemctl start fcfsfused-proxy
```
> fcfsfused-proxy start unix socket under `/var/lib/kubelet/fcfsfused-proxy.sock` by default

 - make sure all required [Protocol Buffers](https://github.com/protocolbuffers/protobuf) binaries are installed
```console
./hack/install-protoc.sh
```
 - when any change is made to `proto/*.proto` file, run below command to generate
```console
rm pkg/fcfsfused-proxy/pb/*.go
protoc --proto_path=pkg/fcfsfused-proxy/proto --go-grpc_out=pkg/fcfsfused-proxy/pb --go_out=pkg/fcfsfused-proxy/pb pkg/fcfsfused-proxy/proto/*.proto
```
 - build new fcfsfused-proxy binary by running
```console
make fcfsfused-proxy
```

 - Generate debian dpkg package
```console
cp _output/fcfsfused-proxy ./pkg/fcfsfused-proxy/debpackage/usr/bin/fcfsfused-proxy
dpkg-deb --build pkg/fcfsfused-proxy/debpackage
```

 - Generate redhat/centos package
```console
cp _output/fcfsfused-proxy ./pkg/fcfsfused-proxy/rpmbuild/SOURCES/fcfsfused-proxy
cd ~/rpmbuild/SPECS/
rpmbuild --target noarch -bb utils.spec
```

- Installing fcfsfused-proxy package
```console
# On debian based systems
wget https://github.com/happyfish100/fastcfs-csi/raw/master/deploy/fcfsfused-proxy/v0.1.0/fcfsfused-proxy-v0.1.0.deb
dpkg -i fcfsfused-proxy-v0.1.0.deb

# On redhat/centos based systems
wget https://github.com/happyfish100/fastcfs-csi/raw/master/deploy/fcfsfused-proxy/v0.1.0/fcfsfused-proxy-v0.1.0.rpm
rpm -ivh utils-1.0.0-1.noarch.rpm
```
