# FastCFS-CSI

[English](./README.md) | 简体中文

## 1. 概述

FastCFS 容器存储接口（CSI）驱动器为容器编排器（Container Orchestrators）管理FastCFS类型卷的生命周期提供 [CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md) 支持。

### 1.1. 项目状态

状态: **beta**

### 1.2. 功能

* **静态供应** - 创建一个新的或迁移现有的 FastCFS 卷, 然后从 FastCFS 卷创建持久卷 (PV) 并使用 PersistentVolumeClaim (PVC) 从容器中消费 PV。
* **动态供应** - 使用PersistentVolumeClaim (PVC)请求 Kuberenetes 代表用户创建 FastCFS 卷，并从容器内部消费该卷。
* **挂载选项** - 可以通过在持久卷 (PV) 中指定挂载选项，来定义卷的挂载方式。
* **[卷扩充](https://kubernetes-csi.github.io/docs/volume-expansion.html)** - 扩充卷的大小。自 Kubernetes 1.16 起，这个 CSI 功能（`ExpandCSIVolumes`）为 `beta` 版。

**注意** FastCFS-CSI 不支持删除静态卷。PV 规范中的 `persistentVolumeReclaimPolicy` 必须设置为 `Retain`，以避免在 csi-provisioner 中尝试删除 PV。

## 2. Kubernetes 上使用 FastCFS CSI 驱动

以下部分是特定于 Kubernetes 的。如果您是 Kubernetes 用户，请使用以下驱动程序功能、安装步骤和示例。

### 2.1. Kubernetes 版本兼容性矩阵
| FastCFS CSI Driver \ Kubernetes Version| v1.17 | v1.18+ |
|----------------------------------------|-------|-------|
| master branch                          | ?     | yes   |
| v0.2.0                                 | ?     | yes   |
| v0.1.0                                 | ?     | yes   |


### 2.2. 先决条件
* [FastCFS](https://github.com/happyfish100/FastCFS/) `v2.1.+ `
* FastCFS 启用验证模块`auth_enabled = true`
* FastCFS 客户端、auth 模块和 cluster 等配置文件需要提供 HTTP 或 HTTPS 访问方式。可以把 FastCFS 所有配置文件拷贝到 web server 的根目录下:
    * 例如：`cp -R /etc/fastcfs /path/www && cd /path/www && python3 -m http.server 8080`.
* 熟悉如何设置 Kubernetes 并拥有一个可以工作的 Kubernetes 集群:
    * `kubelet` 和 `kube-apiserver` 需启用标志 `--allow-privileged=true`
    * 启用 `kube-apiserver` 特性门控 `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true`
    * 启用 `kubelet` 特性门控 `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true`
    
### 2.3. 安装

#### 2.3.1. 设置驱动权限

驱动程序需要 FastCFS 密钥才能与 FastCFS 通信以代表用户管理卷。有一种授予驱动程序权限的方法：

* 使用secret对象 - 在FastCFS创建具有适当权限的管理员和用户，将该用户的凭据（FastCFS的凭证默认目录为`/etc/fastcfs/auth/keys`）放入 [密钥清单](../deploy/kubernetes/secret.yaml), 然后部署。

```sh
curl https://raw.githubusercontent.com/happyfish100/fastcfs-csi/master/deploy/kubernetes/secret.yaml > secret.yaml
# 编辑这个文件，填入你的用户凭证
kubectl apply -f secret.yaml
```

然后就可以在存储类中使用这个密钥

#### 2.3.2. 配置节点容忍（toleration）设置

默认情况下，驱动程序容忍污点 `CriticalAddonsOnly` 并将 `tolerationSeconds` 配置为 `300`，要在任何节点上部署驱动程序，请在部署前将 helm `Value.node.tolerateAllTaints` 设置为 true

#### 2.3.3. 部署驱动

在部署驱动程序之前，请参阅上面的兼容性矩阵

```sh
kubectl apply -k "github.com/happyfish100/fastcfs-csi/deploy/kubernetes/overlays/dev/?ref=main"
```

修改ConfigMap, 并替换它。[ConfigMap 例子](./examples/kubernetes/config-map/README.md)

```sh
curl https://raw.githubusercontent.com/happyfish100/fastcfs-csi/master/deploy/kubernetes/base/csiplugin-configmap.yaml > csiplugin-configmap.yaml
kubectl replace -f csiplugin-configmap.yaml
```

验证驱动程序正在运行:

```sh
kubectl get pods
```


或者，您也可以使用 helm 安装驱动程序：

添加 fastcfs-csi Helm 存储库：
```sh
helm repo add fastcfs-csi https://happyfish100.github.io/fastcfs-csi
helm repo update
```

然后使用 chart 安装驱动程序的版本
```sh
helm upgrade --install fastcfs-csi fastcfs-csi/fcfs-csi-driver
```


#### 2.3.4. 使用调试模式部署驱动程序

要查看驱动程序调试日志，请使用 `-v=5` 命令行选项运行 CSI 驱动程序

### 2.4. 例子

确保在示例之前遵循 [先决条件](README-zh_CN.md#先决条件) :

* [Config Map](./examples/kubernetes/config-map)
* [动态供应](./examples/kubernetes/dynamic-provisioning)
* [静态供应](./examples/kubernetes/static-provisioning)
* [配置存储类](./examples/kubernetes/storageclass)
* [卷扩充](./examples/kubernetes/resizing)

### 2.5. CSI 规范和 Kubernetes 版本兼容性

请参考Kubernetes文档中的 [兼容矩阵](https://kubernetes-csi.github.io/docs/#kubernetes-releases)

## 3. 开发

开发前请先阅读 [CSI Spec](https://github.com/container-storage-interface/spec/blob/master/spec.md) 和 [General CSI driver development guideline](https://kubernetes-csi.github.io/docs/developing.html) 获得对CSI驱动有一些基本的了解。

### 3.1. 要求

* Golang 1.15.+
* [Ginkgo](https://github.com/onsi/ginkgo) 在您的 环境变量 中进行端到端测试
* Docker 17.05+ 发布版

### 3.2. 依赖

通过 go module 管理依赖。要构建项目，首先使用`export GO111MODULE=on`打开go mod，然后使用：`make`构建项目

### 3.3. 测试

* 执行e2e测试，运行：`make test-e2e-single-nn`和`make test-e2e-multi-nn`（现在只能本地执行，本地需要可以连接FastCFS集群）

### 3.4. 构建容器镜像

* 构建镜像 : `make image-csi`

### 3.5. Helm 和 manifests

helm chart 位于 `charts/fcfs-csi-driver` 目录中。manifests 位于 `deploy/kubernetes` 目录中。
除了 kustomize patches 之外的所有清单都是通过运行 `helm template` 生成的。这使 helm chart 和 manifests 保持同步。

更新 helm chart:

* 生成 manifests: `make generate-kustomize`
* `deploy/kubernetes/values` 中有用于生成一些 manifests 的值文件
* 向 helm chart 添加新资源模板时，请更新 `generate-kustomize` 的 make 目标和 `deploy/kubernetes/values` 文件和适当的 kustomization.yaml 文件。
