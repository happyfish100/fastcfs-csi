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

package fcfs

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/klog/v2"
    "os"
)

type Metadata struct {
    labels map[string]string
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
        labels: node.GetLabels(),
    }

    return &metadata, nil

}
