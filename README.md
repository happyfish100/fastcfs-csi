# FastCFS-CSI

English | [简体中文](./README-zh_CN.md)

## Overview

The [FastCFS](https://github.com/happyfish100/FastCFS/) Container Storage Interface (CSI) Driver provides a [CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md) interface used by Container Orchestrators to manage the lifecycle of FastCFS volumes.

## Project status

Status: **beta**

## Features
* **Static Provisioning** - create a new or migrating existing FastCFS volumes, then create persistence volume (PV) from the FastCFS volume and consume the PV from container using persistence volume claim (PVC).
* **Dynamic Provisioning** - uses persistence volume claim (PVC) to request the Kuberenetes to create the FastCFS volume on behalf of user and consumes the volume from inside container. 
* **Mount Option** - mount options could be specified in persistence volume (PV) to define how the volume should be mounted.
* **[Volume Resizing](https://kubernetes-csi.github.io/docs/volume-expansion.html)** - expand the volume size. The corresponding CSI feature (`ExpandCSIVolumes`) is beta since Kubernetes 1.16.

**Note** fastcfs-csi does not supports deletion for static PV.
`persistentVolumeReclaimPolicy` in PV spec must be set to `Retain` to avoid PV delete attempt in csi-provisioner.

# FastCFS CSI Driver on Kubernetes
Following sections are Kubernetes specific. If you are Kubernetes user, use followings for driver features, installation steps and examples.

## FastCFS Version Compatibility Matrix
| FastCFS CSI Driver       | FastCFS Version|
|--------------------------|----------------|
| v0.4.6-fastcfs5.0.0-1    | v5.0.0         | 
| v0.4.6-fastcfs4.2.0-1    | v4.2.0         |  
| v0.4.6-1fastcfs4.1.0     | v4.1.0         |       
| v0.4.6-3fastcfs4.0.0     | v4.0.0         |       
| v0.4.6-1fastcfs3.7.1     | v3.7.1         |   
| v0.4.5                   | v3.6.3         |
| v0.4.3                   | v3.6.0         |
| v0.4.2                   | v3.5.0         |
| v0.4.0                   | v3.3.0         |

## Kubernetes Version Compatibility Matrix
| FastCFS CSI Driver \ Kubernetes Version| v1.17 | v1.18+ |
|----------------------------------------|-------|-------|
| master branch                          | ?     | yes   |
| v0.4.0                                 | ?     | yes   |

## Prerequisites
* [FastCFS](https://github.com/happyfish100/FastCFS/) `v2.2.0 ` or later
* Enable FastCFS Auth `auth_enabled = true`
* Get yourself familiar with how to setup Kubernetes and have a working Kubernetes cluster:
    * Enable flag `--allow-privileged=true` for `kubelet` and `kube-apiserver`
    * Enable `kube-apiserver` feature gates `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true`
    * Enable `kubelet` feature gates `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true`
    
## Installation
#### Set up driver permission
The driver requires FastCFS secret to talk to FastCFS to manage the volume on user's behalf. There is a method to grant driver permission:

- Using secret object - create an admin and user with proper permission, put that user's credentials （The default directory of FastCFS credentials is `/etc/fastcfs/auth/keys`） in [secret manifest](./deploy/kubernetes/secret.yaml) then deploy the secret.

```sh
curl https://raw.githubusercontent.com/happyfish100/fastcfs-csi/master/deploy/kubernetes/secret.yaml > secret.yaml
# Edit the secret with user credentials
kubectl apply -f secret.yaml
```

Then reference this key in your storage class.

#### Config node toleration settings
By default, driver tolerates taint `CriticalAddonsOnly` and has `tolerationSeconds` configured as `300`, to deploy the driver on any nodes, please set helm `Value.node.tolerateAllTaints` to true before deployment

#### Deploy driver
Please see the compatibility matrix above before you deploy the driver

##### deploy

Deploy using kustomization file

```sh
kubectl apply -k "github.com/happyfish100/fastcfs-csi/deploy/kubernetes/overlays/dev/?ref=main"
```

Alternatively, you could also install the driver using helm:

Add the fastcfs-csi Helm repository:
```sh
helm repo add fastcfs-csi https://happyfish100.github.io/fastcfs-csi
helm repo update
```

Then install a release of the driver using the chart
```sh
helm upgrade --install fastcfs-csi fastcfs-csi/fcfs-csi-driver
```

##### edit FastCFS Config
Update FastCFS Config, modify host(`data.fdir-cluster.host`, `data.fstore-cluster.host`, `data.auth-cluster.host`) or more configuration.
For more, see [FastCFS Config Example](./examples/kubernetes/fastcfs-config/README.md)

```sh
kubectl edit configmap fastcfs-client-config
```

#### Verify driver is running:
```sh
kubectl get pods
```


#### Deploy driver with debug mode
To view driver debug logs, run the CSI driver with `-v=5` command line option

## Examples
Make sure you follow the [Prerequisites](README.md#Prerequisites) before the examples:
* [FastCFS Config](./examples/kubernetes/fastcfs-config)
* [Dynamic Provisioning](./examples/kubernetes/dynamic-provisioning)
* [Static Provisioning](./examples/kubernetes/static-provisioning)
* [Configure StorageClass](./examples/kubernetes/storageclass)
* [Volume Resizing](./examples/kubernetes/resizing)

### CSI spec and Kubernetes version compatibility

Please refer to the [matrix](https://kubernetes-csi.github.io/docs/#kubernetes-releases)
in the Kubernetes documentation.

## Development
Please go through [CSI Spec](https://github.com/container-storage-interface/spec/blob/master/spec.md) and [General CSI driver development guideline](https://kubernetes-csi.github.io/docs/developing.html) to get some basic understanding of CSI driver before you start.

### Requirements
* Golang 1.15.+
* [Ginkgo](https://github.com/onsi/ginkgo) in your PATH for end-to-end testing
* Docker 17.05+ for releasing

### Dependency
Dependencies are managed through go module. To build the project, first turn on go mod using `export GO111MODULE=on`, then build the project using: `make`

### Testing

* To execute e2e tests, run: `make test-e2e-single-nn` and `make test-e2e-multi-nn` (Now it can only be executed locally, and you can connect to the FastCFS cluster locally)

### Build Container Image
* Build image : `make image-csi`

### Helm and manifests
The helm chart for this project is in the `charts/fcfs-csi-driver` directory.  The manifests for this project are in the `deploy/kubernetes` directory.  All the manifests except kustomize patches are generated by running `helm template`.  This keeps the helm chart and the manifests in sync.

When updating the helm chart:
* Generate manifests: `make generate-kustomize`
* There are values files in `deploy/kubernetes/values` used for generating some manifests
* When adding a new resource template to the helm chart please update the `generate-kustomize` make target, the `deploy/kubernetes/values` files, and the appropriate kustomization.yaml file(s).
