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

package e2e

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	"k8s.io/kubernetes/test/e2e/framework/volume"
	"strings"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	"vazmin.github.io/fastcfs-csi/tests/e2e/driver"
	"vazmin.github.io/fastcfs-csi/tests/e2e/testsuites"
)

const (
	defaultvolSize = 1

	dummyVolumeName = "pre-provisioned"
)

var (
	defaultvolSizeBytes int64 = volume.GiB
)

func execCommandInPod(f *framework.Framework, c, ns string, opt *metav1.ListOptions) (string, string, error) {
	podPot, err := getCommandInPodOpts(f, c, ns, opt)
	if err != nil {
		return "", "", err
	}
	stdOut, stdErr, err := f.ExecWithOptions(podPot)
	if stdErr != "" {
		e2elog.Logf("stdErr occurred: %v", stdErr)
	}
	return stdOut, stdErr, err
}
func getCommandInPodOpts(f *framework.Framework, c, ns string, opt *metav1.ListOptions) (framework.ExecOptions, error) {
	cmd := []string{"/bin/sh", "-c", c}
	framework.Logf("namespace %q", ns)
	podList, err := f.PodClientNS(ns).List(context.TODO(), *opt)
	framework.ExpectNoError(err)
	if len(podList.Items) == 0 {
		return framework.ExecOptions{}, errors.New("podlist is empty")
	}
	if err != nil {
		return framework.ExecOptions{}, err
	}
	return framework.ExecOptions{
		Command:            cmd,
		PodName:            podList.Items[0].Name,
		Namespace:          ns,
		ContainerName:      podList.Items[0].Spec.Containers[0].Name,
		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: true,
	}, nil
}

// Requires env AVAILABILITY_NODE_NAME a comma separated list of AZs to be set
var _ = Describe("[fcfs-csi-e2e] [single-nn] Pre-Provisioned", func() {
	f := framework.NewDefaultFramework("fcfs")

	var (
		cs         clientset.Interface
		ns         *v1.Namespace
		fcfsDriver driver.PreProvisionedVolumeTestDriver

		//cfs        fcfs.Cfs
		volumeID string
		volSize  string
		//credentials *common.Credentials
		//volOptions *fcfs.VolumeOptions
		cleanupFuncs []func()
		tmpKeyFile   string
	)

	BeforeEach(func() {
		cs = f.ClientSet
		ns = f.Namespace
		fcfsDriver = driver.InitFcfsCSIDriver()

		// setup FastCFS volume
		//if os.Getenv(availabilityNodeNameEnv) == "" {
		//	Skip(fmt.Sprintf("env %q not set", availabilityNodeNameEnv))
		//}
		var err error
		cleanupFuncs, err = testsuites.SetupBaseEnv(cs, ns)
		if err != nil {
			Fail(fmt.Sprintf("could not SetupBaseEnv: %v", err))
		}
		opt := metav1.ListOptions{
			LabelSelector: "app=fcfs-csi-controller",
		}
		volumeID = fmt.Sprintf("csi-static-vol-%s", uuid.New().String())
		tmpKeyFile = fmt.Sprintf("/tmp/csi/keys/%s.key", ns.Name)
		if len(driver.TestOptions.AdminSecretKey) == 0 {
			Fail("admin key must be set")
		}
		_, stderr, err := execCommandInPod(f, fmt.Sprintf("echo %s > %s", driver.TestOptions.AdminSecretKey, tmpKeyFile), driver.DefaultCSINamespace, &opt)
		if err != nil {
			Fail(err.Error())
		}
		if len(stderr) > 0 {
			Fail(stderr)
		}
		args := []string{
			common.PoolCMD,
			"-u", driver.TestOptions.AdminName,
			"-k", tmpKeyFile,
			"-c", driver.TestOptions.ConfigURL + common.PoolConfigFile,
			"create", volumeID,
			fmt.Sprintf("%dg", defaultvolSize),
		}

		_, stderr, err = execCommandInPod(f, strings.Join(args, " "), driver.DefaultCSINamespace, &opt)

		if err != nil {
			Fail(err.Error())
		}
		if len(stderr) > 0 {
			Fail(stderr)
		}
		volSize = fmt.Sprintf("%dGi", defaultvolSize)
	})

	AfterEach(func() {
		//if !skipManuallyDeletingVolume {
		//	args := []string{
		//		common.PoolCMD,
		//		"-u", driver.TestOptions.AdminName,
		//		"-k", tmpKeyFile,
		//		"-c", driver.TestOptions.ConfigURL + common.PoolConfigFile,
		//		"delete", volumeID,
		//		fmt.Sprintf("%dg", defaultvolSize),
		//	}
		//	opt := &metav1.ListOptions{
		//		LabelSelector: "app=fcfs-csi-controller",
		//	}
		//	_, stderr, err := execCommandInPod(f, strings.Join(args, " "), driver.DefaultCSINamespace, opt)
		//
		//	if err != nil {
		//		Fail(err.Error())
		//	}
		//	if len(stderr) > 0 {
		//		Fail(stderr)
		//	}
		//}
		for _, cleanupFunc := range cleanupFuncs {
			cleanupFunc()
		}
	})

	It("[env] should write and read to a pre-provisioned volume", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						VolumeID:  volumeID,
						ClaimSize: volSize,
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
						},
					},
				},
			},
		}
		test := testsuites.PreProvisionedVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

	It("[env] should use a pre-provisioned volume and mount it as readOnly in a pod", func() {
		pods := []testsuites.PodDetails{
			{
				Cmd: "echo 'hello world' > /mnt/test-1/data && grep 'hello world' /mnt/test-1/data",
				Volumes: []testsuites.VolumeDetails{
					{
						VolumeID:  volumeID,
						ClaimSize: volSize,
						VolumeMount: testsuites.VolumeMountDetails{
							NameGenerate:      "test-volume-",
							MountPathGenerate: "/mnt/test-",
							ReadOnly:          true,
						},
					},
				},
			},
		}
		test := testsuites.PreProvisionedReadOnlyVolumeTest{
			CSIDriver: fcfsDriver,
			Pods:      pods,
		}
		test.Run(cs, ns)
	})

})
