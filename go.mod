module vazmin.github.io/fastcfs-csi

go 1.16

require (
	github.com/container-storage-interface/spec v1.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/kubernetes-csi/csi-lib-utils v0.9.1
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20210331060903-cb1fcc7394e5
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubernetes v1.21.0
	k8s.io/mount-utils v0.0.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/apiserver => k8s.io/apiserver v0.21.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0
	k8s.io/code-generator => k8s.io/code-generator v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.0
	k8s.io/cri-api => k8s.io/cri-api v0.21.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0
	k8s.io/kubectl => k8s.io/kubectl v0.21.0
	k8s.io/kubelet => k8s.io/kubelet v0.21.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0
	k8s.io/metrics => k8s.io/metrics v0.21.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.0
	k8s.io/sample-controller => k8s.io/sample-controller v0.21.0
)
