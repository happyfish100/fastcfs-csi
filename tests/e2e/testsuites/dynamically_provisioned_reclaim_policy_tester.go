/*
Copyright 2018 The Kubernetes Authors.

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

/*
Modifications Copyright 2021 vazmin.
Licensed under the Apache License, Version 2.0.
*/

package testsuites

import (
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
	"vazmin.github.io/fastcfs-csi/tests/e2e/driver"
)

// DynamicallyProvisionedReclaimPolicyTest will provision required PV(s) and PVC(s)
// Testing the correct behavior for different reclaimPolicies
type DynamicallyProvisionedReclaimPolicyTest struct {
	CSIDriver driver.DynamicPVTestDriver
	Volumes   []VolumeDetails
	Cfs       fcfs.Cfs
}

func (t *DynamicallyProvisionedReclaimPolicyTest) Run(client clientset.Interface, namespace *v1.Namespace) {
	for _, volume := range t.Volumes {
		tpvc, _ := volume.SetupDynamicPersistentVolumeClaim(client, namespace, t.CSIDriver)

		// will delete the PVC
		// will also wait for PV to be deleted separately when reclaimPolicy=Retian
		tpvc.Cleanup()
		// first check PV stills exists, then manually delete it
		if tpvc.ReclaimPolicy() == v1.PersistentVolumeReclaimRetain {
			tpvc.WaitForPersistentVolumePhase(v1.VolumeReleased)
			tpvc.DeleteBoundPersistentVolume()
			//tpvc.DeleteBackingVolume(t.Cfs)
		}
	}
}
