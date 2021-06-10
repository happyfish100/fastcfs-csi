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

package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
	"vazmin.github.io/fastcfs-csi/tests/e2e/driver"
	"vazmin.github.io/fastcfs-csi/tests/e2e/testsuites"
)

var _ = Describe("[fcfs-csi-e2e] [single-nn] Dynamic Provisioning", func() {
	f := framework.NewDefaultFramework("fcfs")

	var (
		cs           clientset.Interface
		ns           *v1.Namespace
		fcfsDriver   driver.PVTestDriver
		cleanupFuncs []func()
	)

	BeforeEach(func() {
		cs = f.ClientSet
		ns = f.Namespace
		fcfsDriver = driver.InitFcfsCSIDriver()
		var err error
		cleanupFuncs, err = testsuites.SetupBaseEnv(cs, ns)
		if err != nil {
			Fail(fmt.Sprintf("could not setup base env: %v", err))
		}
	})

	AfterEach(func() {
		for _, cleanupFunc := range cleanupFuncs {
			cleanupFunc()
		}
	})

	It("should create a volume on demand", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedCmdVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It("should create a volume on demand with provided mountOptions", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						MountOptions: []string{"rw"},
						ClaimSize:    driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedCmdVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It("should create multiple PV objects, bind to PVCs and attach all to a single pod", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && echo 'hello world' > /mnt/test-2/data && grep 'hello world' /mnt/test-1/data  && grep 'hello world' /mnt/test-2/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedCmdVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It("should create multiple PV objects, bind to PVCs and attach all to different pods", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{

						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedCmdVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It("should create multiple PV objects, bind to PVCs and attach all to different pods on the same node", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "while true; do echo $(date -u) >> /mnt/test-1/data; sleep 1; done",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
			{
				Cmd: "while true; do echo $(date -u) >> /mnt/test-1/data; sleep 1; done",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedCollocatedPodTest{
			CSIDriver:    fcfsDriver,
			Pods:         pods,
			ColocatePods: true,
		}
		test.Run(cs, ns)
	})

	// Track issue https://github.com/kubernetes/kubernetes/issues/70505
	It("should create a volume on demand and mount it as readOnly in a pod", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "touch /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize: driver.MinimumSizeForVolumeType(),
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
							ReadOnly:          true,
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedReadOnlyVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It(fmt.Sprintf("should delete PV with reclaimPolicy %q", v1.PersistentVolumeReclaimDelete), func() {
		reclaimPolicy := v1.PersistentVolumeReclaimDelete
		volumes := []testsuites.VolumeDetails{
			{
				ClaimSize:     driver.MinimumSizeForVolumeType(),
				ReclaimPolicy: &reclaimPolicy,
			},
		}
		test := testsuites.DynamicallyProvisionedReclaimPolicyTest{
			CSIDriver: fcfsDriver,
			Volumes:   volumes,
		}
		test.Run(cs, ns)
	})

	It(fmt.Sprintf("[env] should retain PV with reclaimPolicy %q", v1.PersistentVolumeReclaimRetain), func() {
		allowedTopologyZones := driver.TestOptions.AllowedTopologyValues
		if len(allowedTopologyZones) == 0 {
			Skip(fmt.Sprintf("%q not set", driver.AvailabilityTopologyValues))
		}
		reclaimPolicy := v1.PersistentVolumeReclaimRetain
		volumes := []testsuites.VolumeDetails{
			{
				ClaimSize:     driver.MinimumSizeForVolumeType(),
				ReclaimPolicy: &reclaimPolicy,
			},
		}
		cfs, err := fcfs.NewCFS()
		if err != nil {
			Fail(fmt.Sprintf("could not get NewCFS: %v", err))
		}

		test := testsuites.DynamicallyProvisionedReclaimPolicyTest{
			CSIDriver: fcfsDriver,
			Volumes:   volumes,
			Cfs:       cfs,
		}
		test.Run(cs, ns)
	})

	It("should create a deployment object, write and read to it, delete the pod and write and read to it again", func() {
		pod := testsuites.PodDetails{
			Cmd: "echo 'hello world' >> /mnt/test-1/data && while true; do sleep 1; done",
			Volumes: []testsuites.VolumeDetails{
				{
					ClaimSize: driver.MinimumSizeForVolumeType(),
					VolumeMount: testsuites.VolumeMountDetails{
						NameGenerate:      "test-volume-",
						MountPathGenerate: "/mnt/test-",
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedDeletePodTest{
			CSIDriver: fcfsDriver,
			Pod:       pod,
			PodCheck: &testsuites.PodExecCheck{
				Cmd:            []string{"cat", "/mnt/test-1/data"},
				ExpectedString: "hello world\nhello world\n", // pod will be restarted so expect to see 2 instances of string
			},
		}
		test.Run(cs, ns)
	})

	It("should create a volume on demand and resize it ", func() {
		allowVolumeExpansion := true
		pod := testsuites.PodDetails{
			Cmd: "echo 'hello world' >> /mnt/test-1/data && grep 'hello world' /mnt/test-1/data && sync",
			Volumes: []testsuites.VolumeDetails{
				{
					ClaimSize: driver.MinimumSizeForVolumeType(),
					VolumeMount: testsuites.VolumeMountDetails{
						NameGenerate:      "test-volume-",
						MountPathGenerate: "/mnt/test-",
					},
					AllowVolumeExpansion: &allowVolumeExpansion,
				},
			},
		}
		test := testsuites.DynamicallyProvisionedResizeVolumeTest{
			CSIDriver: fcfsDriver,
			Pod:       pod,
		}
		test.Run(cs, ns)
	})
})

var _ = Describe("[fcfs-csi-e2e] [multi-nn] Dynamic Provisioning", func() {
	f := framework.NewDefaultFramework("fcfs")

	var (
		cs           clientset.Interface
		ns           *v1.Namespace
		fcfsDriver   driver.DynamicPVTestDriver
		cleanupFuncs []func()
	)

	BeforeEach(func() {
		cs = f.ClientSet
		ns = f.Namespace
		fcfsDriver = driver.InitFcfsCSIDriver()
		var err error
		cleanupFuncs, err = testsuites.SetupBaseEnv(cs, ns)
		if err != nil {
			Fail(fmt.Sprintf("could not setup base env: %v", err))
		}
	})

	AfterEach(func() {
		for _, cleanupFunc := range cleanupFuncs {
			cleanupFunc()
		}
	})

	It("should allow for topology aware volume scheduling", func() {
		volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize:         driver.MinimumSizeForVolumeType(),
						VolumeBindingMode: &volumeBindingMode,
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedTopologyAwareVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	// Requires env AVAILABILITY_NODE_NAME, a comma separated list of AZs
	It("[env] should allow for topology aware volume with specified zone in allowedTopologies", func() {
		allowedTopologyZones := driver.TestOptions.AllowedTopologyValues
		if len(allowedTopologyZones) == 0 {
			Skip(fmt.Sprintf("%q not set", driver.AvailabilityTopologyValues))
		}
		volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						ClaimSize:             driver.MinimumSizeForVolumeType(),
						VolumeBindingMode:     &volumeBindingMode,
						AllowedTopologyValues: allowedTopologyZones,
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.DynamicallyProvisionedTopologyAwareVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})
})
