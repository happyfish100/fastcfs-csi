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
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/mount-utils"
	"os"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	csicommon "vazmin.github.io/fastcfs-csi/pkg/csi-common"
	mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	fcfsFusedEndpoint        string
	enableFcfsFusedProxy     bool
	fcfsFusedProxyConnTimout int
	mounter                  mount.Interface
	volumeLocks              *common.VolumeLocks
	topology                 map[string]string
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if err := common.ValidateNodeStageVolumeRequest(request); err != nil {
		return nil, err
	}

	stagingTargetPath := request.GetStagingTargetPath()
	volumeId := request.GetVolumeId()
	cr, err := common.NewAdminCredentials(request.GetSecrets())
	if err != nil {
		klog.Errorf("failed to retrieve user credentials: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer cr.DeleteCredentials()

	if acquired := ns.volumeLocks.TryAcquire(volumeId); !acquired {
		klog.Error(common.VolumeOperationAlreadyExistsFmt, volumeId)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeId)
	}
	defer ns.volumeLocks.Release(volumeId)

	mnt, err := ns.ensureMountPoint(stagingTargetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not mount target %q: %v", stagingTargetPath, err)
	}
	if mnt {
		klog.V(2).Infof("NodeStageVolume: volume %s is already mounted on %s", volumeId, stagingTargetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	vol, err := newFcfsVolumeFromVolID(volumeId, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "new FcfsVolume err %v", err)
	}
	vol.VolPath = stagingTargetPath
	if ns.enableFcfsFusedProxy {
		_, err = mountFcfsFusedWithProxy(ctx, vol, ns.fcfsFusedEndpoint, ns.fcfsFusedProxyConnTimout, request.Secrets)
	} else {
		err = fuseMount(ctx, vol, cr)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "[FcfsCFS] fuse mount err %v", err)
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if err := common.ValidateNodeUnstageVolumeRequest(req); err != nil {
		return nil, err
	}
	volumeID := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()

	if acquired := ns.volumeLocks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.volumeLocks.Release(volumeID)

	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Error(codes.NotFound, "Targetpath not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if notMnt {
		return &csi.NodeUnstageVolumeResponse{}, nil
		// return nil, status.Errorf(codes.NotFound, "Volume not mounted %s", targetPath)
	}

	//vol, err := newFcfsVolumeFromVolID(volumeID, nil)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	klog.V(2).Infof("NodeUnstageVolume: CleanupMountPoint %s on volumeID(%s)", targetPath, volumeID)
	//err = fuseUnmount(ctx, vol)
	err = mount.CleanupMountPoint(targetPath, ns.mounter, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount staging target %q: %v", targetPath, err)
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	volumeId := req.GetVolumeId()
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	//ephemeralVolume := request.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "true" ||
	//	request.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "" && fc.conf.Ephemeral // Kubernetes 1.15 doesn't have csi.storage.k8s.io/ephemeral.

	if req.GetVolumeCapability().GetBlock() != nil {
		return nil, status.Error(codes.InvalidArgument, "FastCFS doesn't support block access type")
	}

	if acquired := ns.volumeLocks.TryAcquire(volumeId); !acquired {
		klog.Errorf(common.VolumeOperationAlreadyExistsFmt, volumeId)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeId)
	}
	defer ns.volumeLocks.Release(volumeId)

	mnt, err := ns.ensureMountPoint(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not mount target %q: %v", targetPath, err)
	}
	if mnt {
		klog.V(2).Infof("NodePublishVolume: volume %s is already mounted on %s", volumeId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	mountOptions := []string{"bind", "_netdev"}
	mountOptions = common.ConstructMountOptions(mountOptions, req.GetVolumeCapability())
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	}
	err = bindMount(ctx, req.GetStagingTargetPath(), req.GetTargetPath(), mountOptions)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "mount bind err %v", err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if acquired := ns.volumeLocks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeID)
	}
	defer ns.volumeLocks.Release(volumeID)

	klog.V(2).Infof("NodeUnpublishVolume: CleanupMountPoint %s on volumeID(%s)", targetPath, volumeID)

	if err := unmountVolume(ctx, targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("[FastCFS] successfully unbind volume %s from %s", req.GetVolumeId(), targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {

	if len(req.VolumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats volume ID was empty")
	}
	if len(req.VolumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats volume path was empty")
	}

	notMount, err := ns.mounter.IsLikelyNotMountPoint(req.VolumePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "NodeGetVolumeStats: path %s does not exist", req.VolumePath)
		}
		return nil, status.Errorf(codes.Internal, "failed to stat file %s: %v", req.VolumePath, err)
	}
	if notMount {
		return nil, status.Errorf(codes.InvalidArgument, "volume path %s is not mounted", req.VolumePath)
	}

	volumeId := req.GetVolumeId()
	if acquired := ns.volumeLocks.TryAcquire(volumeId); !acquired {
		klog.Errorf(common.VolumeOperationAlreadyExistsFmt, volumeId)
		return nil, status.Errorf(codes.Aborted, common.VolumeOperationAlreadyExistsFmt, volumeId)
	}
	defer ns.volumeLocks.Release(volumeId)

	volumeMetrics, err := volume.NewMetricsStatFS(req.VolumePath).GetMetrics()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get metrics: %v", err)
	}

	available, ok := volumeMetrics.Available.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume available size(%v)", volumeMetrics.Available)
	}
	capacity, ok := volumeMetrics.Capacity.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume capacity size(%v)", volumeMetrics.Capacity)
	}
	used, ok := volumeMetrics.Used.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform volume used size(%v)", volumeMetrics.Used)
	}

	inodesFree, ok := volumeMetrics.InodesFree.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes free(%v)", volumeMetrics.InodesFree)
	}
	inodes, ok := volumeMetrics.Inodes.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes(%v)", volumeMetrics.Inodes)
	}
	inodesUsed, ok := volumeMetrics.InodesUsed.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to transform disk inodes used(%v)", volumeMetrics.InodesUsed)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Unit:      csi.VolumeUsage_BYTES,
				Available: available,
				Total:     capacity,
				Used:      used,
			},
			{
				Unit:      csi.VolumeUsage_INODES,
				Available: inodesFree,
				Total:     inodes,
				Used:      inodesUsed,
			},
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume")
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, request *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            ns.Driver.NodeID,
		MaxVolumesPerNode: ns.Driver.MaxVolumesPerNode,
		AccessibleTopology: &csi.Topology{
			Segments: ns.topology,
		},
	}, nil
}

// ensureMountPoint: create mount point if not exists
// return <true, nil> if it's already a mounted point otherwise return <false, nil>
func (ns *nodeServer) ensureMountPoint(target string) (bool, error) {
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(target)
	if err != nil && !os.IsNotExist(err) {
		if IsCorruptedDir(target) {
			notMnt = false
			klog.Warningf("detected corrupted mount for targetPath [%s]", target)
		} else {
			return !notMnt, err
		}
	}

	if !notMnt {
		// testing original mount point, make sure the mount link is valid
		_, err := ioutil.ReadDir(target)
		if err == nil {
			klog.V(2).Infof("already mounted to target %s", target)
			return !notMnt, nil
		}
		// mount link is invalid, now unmount and remount later
		klog.Warningf("ReadDir %s failed with %v, unmount this directory", target, err)
		if err := ns.mounter.Unmount(target); err != nil {
			klog.Errorf("Unmount directory %s failed with %v", target, err)
			return !notMnt, err
		}
		notMnt = true
		return !notMnt, err
	}
	if err := common.MakeDir(target); err != nil {
		klog.Errorf("MakeDir failed on target: %s (%v)", target, err)
		return !notMnt, err
	}
	return !notMnt, nil
}

func IsCorruptedDir(dir string) bool {
	_, pathErr := mount.PathExists(dir)
	return pathErr != nil && mount.IsCorruptedMnt(pathErr)
}

type MountClient struct {
	service mount_fcfs_fused.MountServiceClient
}

// NewMountClient returns a new mount client
func NewMountClient(cc *grpc.ClientConn) *MountClient {
	service := mount_fcfs_fused.NewMountServiceClient(cc)
	return &MountClient{service}
}
