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

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	CsiConfigFile  = "/etc/fcfs-csi-config/config.json"
	ClientBasePath = "/opt/fastcfs"
	PidSuffixPath  = "fused.pid"
)

type Config struct {
	Endpoint          string // CSI endpoint
	DriverName        string // name of the driver
	NodeID            string // node id
	Ephemeral         bool   // publish volumes in ephemeral mode even if kubelet did not ask for it (only needed for Kubernetes 1.15)
	MaxVolumesPerNode int64  // limit of volumes per node
	Version           bool   // Show version
	DomainLabels      []string

	IsControllerServer bool
	IsNodeServer       bool

	FcfsFusedProxyEndpoint   string
	EnableFcfsFusedProxy     bool
	FcfsFusedProxyConnTimout int
}

type ClusterInfo struct {
	ClusterID string `json:"clusterID"`
	ConfigURL string `json:"configURL"`
}

func readClusterInfo(pathToConfig, clusterID string) (*ClusterInfo, error) {
	var config []ClusterInfo

	// #nosec
	content, err := ioutil.ReadFile(pathToConfig)
	if err != nil {
		err = fmt.Errorf("error fetching configuration for cluster ID %q: %w", clusterID, err)
		return nil, err
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed (%w), raw buffer response: %s",
			err, string(content))
	}

	for _, cluster := range config {
		if cluster.ClusterID == clusterID {
			return &cluster, nil
		}
	}

	return nil, fmt.Errorf("missing configuration for cluster ID %q", clusterID)
}

func ConfigURL(pathToConfig, clusterID string) (string, error) {
	cluster, err := readClusterInfo(pathToConfig, clusterID)
	if err != nil {
		return "", err
	}

	if len(cluster.ConfigURL) == 0 {
		return "", fmt.Errorf("empty config file for cluster ID (%s) in config", clusterID)
	}
	return cluster.ConfigURL, nil
}
