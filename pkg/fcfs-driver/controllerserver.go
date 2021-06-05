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
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	csicommon "vazmin.github.io/fastcfs-csi/pkg/csi-common"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
)

var (
	// NewCFSFunc is a variable for the cloud.NewCloud function that can
	// be overwritten in unit tests.
	NewCFSFunc = fcfs.NewCFS
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	cfs         fcfs.Cfs
	volumeLocks *common.VolumeLocks
}

func NewControllerServer(d *csicommon.CSIDriver) (*controllerServer, error) {
	cfsSrv, _ := NewCFSFunc()

	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		volumeLocks:             common.NewVolumeLocks(),
		cfs:                     cfsSrv,
	}, nil
}

// CreateVolume create volume
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, finalErr error) {
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.V(3).Infof("invalid create FcfsVolume req: %v", req)
		return nil, err
	}

	requestName := req.GetName()

	cr, err := common.NewAdminCredentials(req.GetSecrets())
	if err != nil {
		klog.Errorf("failed to retrieve admin credentials: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer cr.DeleteCredentials()

	// Check arguments
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	// can only mount
	for _, cap := range caps {
		if cap.GetBlock() != nil {
			return nil, status.Error(codes.InvalidArgument, "cannot have block access type")
		}
	}

	// Existence and conflict checks
	if acquired := cs.volumeLocks.TryAcquire(requestName); !acquired {
		klog.Errorf(common.VolumeOperationAlreadyExistsFmt, requestName)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, requestName)
	}
	defer cs.volumeLocks.Release(requestName)

	volOptions, err := newVolumeOptions(ctx, req, requestName, cr)
	if err != nil {
		klog.Errorf("validation and extraction of FcfsVolume options failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	exists, err := cs.cfs.VolumeExists(ctx, volOptions.BaseConfigURL, volOptions.VolName, cr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create FcfsVolume %v: %q", volOptions.VolID, err)
	}
	// TODO if exists to check capacity
	//	if exVol.VolSize < capacity {
	//		return nil, status.Errorf(codes.AlreadyExists, "Volume with the same name: %s but with different size already exist", req.GetName())
	//	}
	if !exists {
		_, createErr := cs.cfs.CreateVolume(ctx, volOptions, cr)
		if createErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to create FcfsVolume %v: %q", volOptions.VolID, createErr)
		}
		klog.V(4).Infof("created FcfsVolume %s at path %s", volOptions.VolID, volOptions.VolPath)
	}

	// VolumeContentSource. Not yet supported VolumeSnapshot and PersistentVolumeClaim Cloning

	csiVol := &csi.Volume{
		VolumeId:      volOptions.VolID,
		CapacityBytes: common.RoundOffBytes(req.GetCapacityRange().GetRequiredBytes()),
		VolumeContext: req.GetParameters(),
		ContentSource: req.GetVolumeContentSource(),
	}
	topologies := common.GetTopologyFromParams(req.GetParameters(), req.GetAccessibilityRequirements())
	if topologies != nil {
		csiVol.AccessibleTopology = []*csi.Topology{
			{
				Segments: topologies,
			},
		}
	}

	return &csi.CreateVolumeResponse{
		Volume: csiVol,
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.V(3).Infof("invalid delete FcfsVolume req: %v", req)
		return nil, err
	}

	cr, err := common.NewAdminCredentials(req.GetSecrets())
	if err != nil {
		klog.Errorf("failed to retrieve admin credentials: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer cr.DeleteCredentials()

	volID := req.GetVolumeId()

	if acquired := cs.volumeLocks.TryAcquire(volID); !acquired {
		klog.Errorf(common.VolumeOperationAlreadyExistsFmt, volID)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volID)
	}
	defer cs.volumeLocks.Release(volID)

	vol, err := newVolOptionsFromVolID(volID, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := cs.cfs.DeleteVolume(ctx, vol, cr); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete FcfsVolume %v: %v", volID, err)
	}
	klog.V(4).Infof("FcfsVolume %v successfully deleted", volID)

	return &csi.DeleteVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks whether the volume capabilities requested are supported.
func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeId := req.GetVolumeId()

	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}
	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Volume %s Capabilities cannot be empty", volumeId)
	}

	for _, capability := range req.GetVolumeCapabilities() {
		if capability.GetBlock() != nil {
			return nil, status.Error(codes.InvalidArgument, "FastCFS doesn't support Block volume")
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}, nil
}

//func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
//	//
//	return nil, status.Error(codes.Unimplemented, "GetCapacity")
//}

func (cs *controllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	// Secrets not in ControllerGetVolumeRequest
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume")
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	// validate
	volumeId := req.GetVolumeId()
	secrets := req.GetSecrets()

	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if acquired := cs.volumeLocks.TryAcquire(volumeId); !acquired {
		klog.Errorf(common.VolumeOperationAlreadyExistsFmt, volumeId)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeId)
	}
	defer cs.volumeLocks.Release(volumeId)

	cr, err := common.NewAdminCredentials(secrets)
	if err != nil {
		klog.Errorf("failed to retrieve admin credentials: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer cr.DeleteCredentials()

	vol, err := newVolOptionsFromVolID(volumeId, req.GetCapacityRange())
	if err != nil {
		klog.Errorf("failed to new volume %s: %v", volumeId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	newSize, err := cs.cfs.ResizeVolume(ctx, vol, cr)
	if err != nil {
		klog.Errorf("failed to expand volume %s: %v", volumeId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: false,
	}, nil
}

func (cs *controllerServer) validateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range cs.Driver.Cap {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported capability %s", c)
}
