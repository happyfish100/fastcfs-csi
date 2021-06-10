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
	"k8s.io/klog/v2"
	"os"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	fcfs "vazmin.github.io/fastcfs-csi/pkg/fcfs-driver"
)

var (
	conf   common.Config
	osExit = os.Exit
)

func initFlag() {

	flag.StringVar(&conf.Endpoint, "endpoint", common.DefaultCSIEndpoint, "CSI endpoint")
	flag.StringVar(&conf.DriverName, "driver-name", common.DefaultDriverName, "name of the driver")
	flag.StringVar(&conf.NodeID, "nodeid", "", "node id")
	flag.BoolVar(&conf.Ephemeral, "ephemeral", false, "publish volumes in ephemeral mode even if kubelet did not ask for it (only needed for Kubernetes 1.15)")
	flag.Int64Var(&conf.MaxVolumesPerNode, "max-volumes-per-node", 0, "limit of volumes per node")
	flag.BoolVar(&conf.Version, "version", false, "Show version.")
	flag.Var(common.NewStringSlice(&conf.DomainLabels), "domain-labels", "topology")

	flag.BoolVar(&conf.IsNodeServer, "node-server", false, "start fastcfs-csi node server")
	flag.BoolVar(&conf.IsControllerServer, "controller-server", false, "start fastcfs-csi controller server")

	flag.StringVar(&conf.FcfsFusedProxyEndpoint, "fcfsfused-proxy-endpoint", "unix://tmp/fcfsfused-proxy.sock", "fcfsfused-proxy endpoint")
	flag.BoolVar(&conf.EnableFcfsFusedProxy, "enable-fcfsfused-proxy", false, "enable fcfsfused-proxy")
	flag.IntVar(&conf.FcfsFusedProxyConnTimout, "fcfsfused-proxy-conn-timeout", 5, "fcfsfused proxy connection timeout(seconds)")

	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		klog.Exitf("failed to set logtostderr flag: %v", err)
	}
	flag.Parse()
}

func main() {
	initFlag()
	if conf.Version {
		info, err := common.GetVersionJSON()
		if err != nil {
			klog.Fatalln(err)
		}
		fmt.Println(info)
		osExit(0)
	}

	driver := fcfs.NewFcfsDriver()

	driver.Run(&conf)
	os.Exit(1)
}
