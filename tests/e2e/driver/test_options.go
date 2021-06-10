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
	"flag"
	"log"
	"os"
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

type testOptions struct {
	TopologyKey           string
	AllowedTopologyValues []string
	ClusterID             string
	ConfigURL             string

	SecretName     string
	AdminName      string
	AdminSecretKey string
	UserName       string
	UserSecretKey  string

	ContainerName     string
	ControllerPodName string
}

var TestOptions testOptions

func RegisterCommonFlags(flags *flag.FlagSet) {
	flags.StringVar(&TestOptions.SecretName, "secret-name", "csi-fcfs-secret", "")
	flags.StringVar(&TestOptions.TopologyKey, "topology-key", "topology.fcfs.csi.vazmin.github.io/hostname", "")
	flags.StringVar(&TestOptions.ClusterID, "cluster-id", "virtual-cluster-id-1", "virtual cluster id")
	flags.StringVar(&TestOptions.ConfigURL, "config-url", "", "fastcfs config base URL")
	flags.StringVar(&TestOptions.AdminName, "admin-name", os.Getenv("FCFS_ADMIN_NAME"), "fastcfs admin name")
	flags.StringVar(&TestOptions.UserName, "username", os.Getenv("FCFS_USER_NAME"), "fastcfs user name")
	flags.StringVar(&TestOptions.ContainerName, "container-name", "", "container name")
	flags.StringVar(&TestOptions.ControllerPodName, "controller-pod-name", "", "controller pod name")
	flags.Var(common.NewStringSlice(&TestOptions.AllowedTopologyValues), AvailabilityTopologyValues, "allowed topology values")
}

func (t *testOptions) Verify() {
	t.AdminSecretKey = os.Getenv("FCFS_ADMIN_SERRET_KEY")
	if len(t.AdminSecretKey) == 0 {
		log.Fatal("env FCFS_ADMIN_SERRET_KEY is required.")
	}
	t.UserSecretKey = os.Getenv("FCFS_USER_SERRET_KEY")
	if len(t.UserSecretKey) == 0 {
		log.Fatal("env FCFS_USER_SERRET_KEY is required.")
	}
	if len(t.ConfigURL) == 0 {
		log.Fatal("-config-url is required.")
	}
	if len(t.UserName) == 0 {
		log.Fatal("-username is required.")
	}
	if len(t.ContainerName) == 0 {
		log.Fatal("-container-name is required.")
	}
	if len(t.ControllerPodName) == 0 {
		log.Fatal("-controller-pod-name is required.")
	}
}
