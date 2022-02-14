/*
Copyright 2021 vazmin.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)

type MetadataService interface {
	GetLabels() map[string]string
	GetAttachedPVOnNode(context context.Context, nodeName string, driverName string) ([]*corev1.PersistentVolume, error)
	GetPodsOnNode(context context.Context, nodeName string) (*corev1.PodList, error)
	GetStoreClassOnDriver(ctx context.Context, driverName string) ([]*storagev1.StorageClass, error)
	GetCredentials(ctx context.Context, ref *corev1.SecretReference) (map[string]string, error)
}

type Metadata struct {
	clientset kubernetes.Interface
	labels    map[string]string
}

var _ MetadataService = &Metadata{}

func (m *Metadata) GetLabels() map[string]string {
	return m.labels
}

func NewMetadata(nodeName string) (MetadataService, error) {
	configPath := os.Getenv("KUBERNETES_CONFIG_PATH")
	var err error
	var config *rest.Config
	if configPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			klog.Fatalf("Failed to get cluster config with error: %v\n", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Failed to get cluster config with error: %v\n", err)
		}
	}
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	metadataService, err := NewMetadataService(nodeName, clientset)
	if err != nil {
		return nil, fmt.Errorf("error getting information from metadata service or node object: %w", err)
	}
	return metadataService, err
}

// NewMetadataService returns a new MetadataServiceImplementation.
func NewMetadataService(nodeName string, clientset kubernetes.Interface) (MetadataService, error) {

	if nodeName == "" {
		return nil, fmt.Errorf("instance metadata is unavailable and CSI_NODE_NAME env var not set")
	}

	// get node with k8s API
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metadata := Metadata{
		clientset: clientset,
		labels:    node.GetLabels(),
	}

	return &metadata, nil

}

// GetAttachedPVOnNode finds all persistent volume objects attached in the node and controlled by csi.
func (m *Metadata) GetAttachedPVOnNode(context context.Context, nodeName string, driverName string) ([]*corev1.PersistentVolume, error) {
	vaList, err := m.clientset.StorageV1().VolumeAttachments().List(context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list VolumeAttachments: %v", err)
	}

	nodePVNames := make(map[string]struct{})
	for _, va := range vaList.Items {
		if va.Spec.NodeName == nodeName &&
			va.Spec.Attacher == driverName &&
			va.Status.Attached &&
			va.Spec.Source.PersistentVolumeName != nil {
			nodePVNames[*va.Spec.Source.PersistentVolumeName] = struct{}{}
		}
	}

	pvList, err := m.clientset.CoreV1().PersistentVolumes().List(context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list PersistentVolumes: %v", err)
	}

	nodePVs := make([]*corev1.PersistentVolume, 0, len(nodePVNames))
	for i := range pvList.Items {
		_, exist := nodePVNames[pvList.Items[i].Name]
		if exist {
			nodePVs = append(nodePVs, &pvList.Items[i])
		}
	}

	return nodePVs, nil
}

func (m *Metadata) GetPodsOnNode(ctx context.Context, nodeName string) (*corev1.PodList, error) {
	return m.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
}

func (m *Metadata) GetStoreClassOnDriver(ctx context.Context, driverName string) ([]*storagev1.StorageClass, error) {
	scl, err := m.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var scItems []*storagev1.StorageClass
	for _, item := range scl.Items {
		if item.Provisioner != driverName {
			continue
		}
		scItems = append(scItems, &item)
	}
	return scItems, nil
}

func (m *Metadata) GetCredentials(ctx context.Context, ref *corev1.SecretReference) (map[string]string, error) {
	if ref == nil {
		return nil, nil
	}

	secret, err := m.clientset.CoreV1().Secrets(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting secret %s in namespace %s: %v", ref.Name, ref.Namespace, err)
	}

	credentials := map[string]string{}
	for key, value := range secret.Data {
		credentials[key] = string(value)
	}
	return credentials, nil
}
