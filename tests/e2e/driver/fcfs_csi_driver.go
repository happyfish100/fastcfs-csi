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
	"fmt"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

const (
	True                       = "true"
	DefaultCSINamespace        = "default"
	AvailabilityTopologyValues = "allowed-topology-values"
)

// Implement DynamicPVTestDriver interface
type fcfsCSIDriver struct {
	driverName string
}

// InitFcfsCSIDriver returns fcfsCSIDriver that implements DynamicPVTestDriver interface
func InitFcfsCSIDriver() PVTestDriver {
	return &fcfsCSIDriver{
		driverName: common.DefaultDriverName,
	}
}

func (d *fcfsCSIDriver) GetDynamicProvisionStorageClass(parameters map[string]string, mountOptions []string, reclaimPolicy *v1.PersistentVolumeReclaimPolicy, volumeExpansion *bool, bindingMode *storagev1.VolumeBindingMode, allowedTopologyValues []string, namespace string) *storagev1.StorageClass {
	provisioner := d.driverName
	generateName := fmt.Sprintf("%s-%s-dynamic-sc-", namespace, provisioner)
	allowedTopologies := []v1.TopologySelectorTerm{}

	if len(allowedTopologyValues) > 0 {
		allowedTopologies = []v1.TopologySelectorTerm{
			{
				MatchLabelExpressions: []v1.TopologySelectorLabelRequirement{
					{
						Key:    TestOptions.TopologyKey,
						Values: allowedTopologyValues,
					},
				},
			},
		}
	}
	return getStorageClass(generateName, provisioner, parameters, mountOptions, reclaimPolicy, volumeExpansion, bindingMode, allowedTopologies)
}

func (d *fcfsCSIDriver) GetPersistentVolume(volumeID string, fsType string, size string, reclaimPolicy *v1.PersistentVolumeReclaimPolicy, namespace string) *v1.PersistentVolume {
	provisioner := d.driverName
	generateName := volumeID
	// Default to Retain ReclaimPolicy for pre-provisioned volumes
	pvReclaimPolicy := v1.PersistentVolumeReclaimRetain
	if reclaimPolicy != nil {
		pvReclaimPolicy = *reclaimPolicy
	}
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    namespace,
			// TODO remove if https://github.com/kubernetes-csi/external-provisioner/issues/202 is fixed
			Annotations: map[string]string{
				"pv.kubernetes.io/provisioned-by": provisioner,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(size),
			},

			PersistentVolumeReclaimPolicy: pvReclaimPolicy,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       provisioner,
					VolumeHandle: volumeID,
					FSType:       fsType,
					NodeStageSecretRef: &v1.SecretReference{
						Name:      TestOptions.SecretName,
						Namespace: namespace,
					},
					VolumeAttributes: map[string]string{
						"clusterID": TestOptions.ClusterID,
						"static":    "true",
					},
				},
			},
		},
	}
}

//func (d *fcfsCSIDriver) GetDynamicProvisionStorageClassToCleanup(parameters map[string]string, namespace string) *storagev1.StorageClass {
//	provisioner := d.driverName
//	generateName := fmt.Sprintf("%s-%s-cleanup-sc-", namespace, provisioner)
//	return getStorageClass(generateName, provisioner, parameters, nil, nil, nil, nil, nil)
//}

// GetParameters returns the parameters specific for this driver
func GetParameters(volumeType string, namespace *v1.Namespace) map[string]string {
	parameters := map[string]string{
		"type": volumeType,
		"csi.storage.k8s.io/provisioner-secret-name":            TestOptions.SecretName,
		"csi.storage.k8s.io/provisioner-secret-namespace":       namespace.Name,
		"csi.storage.k8s.io/controller-expand-secret-name":      TestOptions.SecretName,
		"csi.storage.k8s.io/controller-expand-secret-namespace": namespace.Name,
		"csi.storage.k8s.io/node-stage-secret-name":             TestOptions.SecretName,
		"csi.storage.k8s.io/node-stage-secret-namespace":        namespace.Name,
		"csi.storage.k8s.io/node-publish-secret-name":           TestOptions.SecretName,
		"csi.storage.k8s.io/node-publish-secret-namespace":      namespace.Name,
		"domainLabels": TestOptions.TopologyKey,
		"clusterID":    TestOptions.ClusterID,
	}
	return parameters
}

// MinimumSizeForVolumeType returns the minimum disk size for each volumeType
func MinimumSizeForVolumeType() string {
	return "1Gi"
}
