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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	csicommon "vazmin.github.io/fastcfs-csi/pkg/csi-common"
)

type fcfsDriver struct {
	driver *csicommon.CSIDriver

	ids *identityServer
	cs  *controllerServer
	ns  *nodeServer
}

type accessType int

const (
	mountAccess accessType = iota
	blockAccess
)
const TopologyKeyNode = "topology.fcfs.csi/node"

func NewFcfsDriver() *fcfsDriver {
	return &fcfsDriver{}
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		volumeLocks:             common.NewVolumeLocks(),
	}
}

func NewNodeServer(d *csicommon.CSIDriver, enableFcfsFusedProxy bool, fcfsFusedEndpoint string, fcfsFusedProxyConnTimout int, topology map[string]string) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d, topology),
		mounter: &mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      utilexec.New(),
		},
		volumeLocks: common.NewVolumeLocks(),
		enableFcfsFusedProxy: enableFcfsFusedProxy,
		fcfsFusedEndpoint: fcfsFusedEndpoint,
		fcfsFusedProxyConnTimout: fcfsFusedProxyConnTimout,
	}
}

func (fc *fcfsDriver) Run(conf *common.Config) {
	fc.driver = csicommon.NewCSIDriver(conf.DriverName, common.DriverVersion, conf.NodeID)
	if fc.driver == nil {
		klog.Fatalln("Failed to initialize CSI Driver")
	}

	if conf.IsControllerServer || !conf.IsNodeServer {
		if !conf.Ephemeral {
			fc.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
				csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
			})
		}
		fc.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
			csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		})
	}
	topology := map[string]string{TopologyKeyNode: conf.NodeID}
	both := !conf.IsControllerServer && !conf.IsNodeServer
	fc.ids = NewIdentityServer(fc.driver)
	if conf.IsControllerServer || both {
		fc.cs = NewControllerServer(fc.driver)
	}

	if conf.IsNodeServer || both {
		fc.driver.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
			csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
		})
		fc.ns = NewNodeServer(fc.driver, conf.EnableFcfsFusedProxy, conf.FcfsFusedProxyEndpoint, conf.FcfsFusedProxyConnTimout, topology)
	}

	s := csicommon.NewNonBlockingGRPCServer()

	s.Start(conf.Endpoint, fc.ids, fc.cs, fc.ns, false)
	s.Wait()

}
