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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	csicommon "vazmin.github.io/fastcfs-csi/pkg/csi-common"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
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

// NewMetadataFunc is a variable for the cloud.NewMetadata function that can
// be overwritten in unit tests.
var NewMetadataFunc = fcfs.NewMetadata

func NewFcfsDriver() *fcfsDriver {
	return &fcfsDriver{}
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewNodeServer(d *csicommon.CSIDriver, enableFcfsFusedProxy bool, fcfsFusedEndpoint string, fcfsFusedProxyConnTimout int, topology map[string]string) *nodeServer {
	mountOptions := &fcfs.MountOptions{
		EnableFcfsFusedProxy:     enableFcfsFusedProxy,
		FcfsFusedEndpoint:        fcfsFusedEndpoint,
		FcfsFusedProxyConnTimout: fcfsFusedProxyConnTimout,
	}
	nodeMounter, err := newNodeMounter()
	if err != nil {
		panic(err)
	}
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d, topology),
		mounter:           nodeMounter,
		volumeLocks:       common.NewVolumeLocks(),
		mountOptions:      mountOptions,
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
	metadataSrv, err := NewMetadataFunc(conf.NodeID)
	if err != nil {
		klog.Fatalln("Failed New Metadata, %v, %q", err, conf.NodeID)
	}
	topology, err := common.GetTopologyFromDomainLabels(metadataSrv.GetLabels(), conf.DomainLabels, conf.DriverName)
	if err != nil {
		klog.Fatalln("Failed GetTopologyFromDomainLabels, %v, %q", err, conf.NodeID)
	}
	klog.V(4).Infof("topology form domain labels: %q", topology)

	both := !conf.IsControllerServer && !conf.IsNodeServer
	fc.ids = NewIdentityServer(fc.driver)
	if conf.IsControllerServer || both {
		fc.cs, err = NewControllerServer(fc.driver)
		if err != nil {
			klog.Fatalln("Failed New Controller Server, %v, %q", err, conf.NodeID)
		}
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
