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

package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/happyfish100/fastcfs-csi/pkg/common"
	fcfs "github.com/happyfish100/fastcfs-csi/pkg/fcfs-driver"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	conf common.Config
)

func init() {
	flag.StringVar(&conf.Endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&conf.DriverName, "drivername", "fcfs.csi.vazmin.github.io", "name of the driver")
	flag.StringVar(&conf.NodeID, "nodeid", "", "node id")
	flag.BoolVar(&conf.Ephemeral, "ephemeral", false, "publish volumes in ephemeral mode even if kubelet did not ask for it (only needed for Kubernetes 1.15)")
	flag.Int64Var(&conf.MaxVolumesPerNode, "maxvolumespernode", 0, "limit of volumes per node")
	flag.BoolVar(&conf.Version, "version", false, "Show version.")
	flag.StringVar(&conf.DomainLabels, "domainlabels", "", "topology")

	flag.BoolVar(&conf.IsNodeServer, "node-server", false, "start fastcfs-csi node server")
	flag.BoolVar(&conf.IsControllerServer, "controller-server", false, "start fastcfs-csi controller server")
}

func main() {
	flag.Parse()

	if conf.Version {
		baseName := path.Base(os.Args[0])
		fmt.Println(baseName, common.DriverVersion)
		return
	}

	driver := fcfs.NewFcfsDriver()

	driver.Run(&conf)
	os.Exit(1)
}
