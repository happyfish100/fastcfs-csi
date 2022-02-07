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
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/mount-utils"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	csicommon "vazmin.github.io/fastcfs-csi/pkg/csi-common"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	mountOptions   *fcfs.MountOptions
	mounter        Mounter
	volumeLocks    *common.VolumeLocks
	ms             MetadataService
	kubeletRootDir string
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if err := common.ValidateNodeStageVolumeRequest(request); err != nil {
		return nil, err
	}

	stagingTargetPath := request.GetStagingTargetPath()
	volumeId := request.GetVolumeId()

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

	volOptions, err := NewVolOptionsFromVolIDOrStatic(volumeId, request.GetVolumeContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "New FastCFS Volume Options ERR %v", err)
	}
	volOptions.VolPath = stagingTargetPath

	mountOptions := &fcfs.MountOptionsSecrets{
		MountOptions: ns.mountOptions,
		Secrets:      request.Secrets,
	}

	err = ns.mounter.FcfsMount(ctx, volOptions, mountOptions)

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

	//vol, err := NewVolOptionsFromVolID(volumeID, nil)
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
	stagingTargetPath := req.GetStagingTargetPath()
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "StagingTargetPath path not provided")
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
	//err = bindMount(ctx, req.GetStagingTargetPath(), targetPath, mountOptions)
	if err := ns.mounter.Mount(stagingTargetPath, targetPath, "", mountOptions); err != nil {
		if removeErr := os.Remove(targetPath); removeErr != nil {
			return nil, status.Errorf(codes.Internal, "Could not remove mount target %q: %v", targetPath, removeErr)
		}
		return nil, status.Errorf(codes.Internal, "Could not mount %q at %q: %v", stagingTargetPath, targetPath, err)
	}
	klog.V(4).Infof("successfully mount %s to %s, %v", stagingTargetPath, targetPath, mountOptions)
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
			Segments: ns.Driver.Topology,
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
	if err := ns.mounter.MakeDir(target); err != nil {
		klog.Errorf("MakeDir failed on target: %s (%v)", target, err)
		return !notMnt, err
	}
	return !notMnt, nil
}

type persistentVolumeWithPods struct {
	*corev1.PersistentVolume
	pods        []*corev1.Pod
	credentials map[string]string
}

func (p *persistentVolumeWithPods) appendPodUnique(new *corev1.Pod) {
	for _, old := range p.pods {
		if old.UID == new.UID {
			return
		}
	}

	p.pods = append(p.pods, new)
}

// getAttachedPVWithPodsOnNode finds all persistent volume objects as well as the related pods in the node.
func (ns *nodeServer) getAttachedPVWithPodsOnNode(ctx context.Context, nodeName string) ([]*persistentVolumeWithPods, error) {
	pvs, err := ns.ms.GetAttachedPVOnNode(ctx, nodeName, ns.Driver.Name)
	if err != nil {
		return nil, fmt.Errorf("getAttachedPVOnNode faied: %v", err)
	}
	scl, err := ns.ms.GetStoreClassOnDriver(ctx, ns.Driver.Name)
	if err != nil {
		return nil, fmt.Errorf("GetStoreClassOnDriver failed: %v", err)
	}
	claimedPVWithPods := make(map[string]*persistentVolumeWithPods, len(pvs))
	crs := make(map[string]map[string]string)
	for _, pv := range pvs {
		if pv.Spec.ClaimRef == nil {
			continue
		}
		sr, err := getSecretReferenceByStorageClassOrPV(pv, scl)
		if err != nil {
			klog.Warningf("get secret reference failed: %v", err)
			continue
		}
		if sr == nil {
			continue
		}
		volumeWithPods := &persistentVolumeWithPods{
			PersistentVolume: pv,
		}
		srKey := fmt.Sprintf("%s/%s", sr.Namespace, sr.Name)
		if cr, ok := crs[srKey]; ok {
			volumeWithPods.credentials = cr
		} else {
			volumeWithPods.credentials, err = ns.ms.GetCredentials(ctx, sr)
			if err != nil {
				klog.Warningf("GetCredentials failed: %v", err)
			}
		}
		pvcKey := fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		claimedPVWithPods[pvcKey] = volumeWithPods
	}

	allPodsOnNode, err := ns.ms.GetPodsOnNode(ctx, nodeName)
	if err != nil {
		return nil, fmt.Errorf("list pods failed: %v", err)
	}

	for i := range allPodsOnNode.Items {
		pod := allPodsOnNode.Items[i]

		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim == nil {
				continue
			}
			pvcKey := fmt.Sprintf("%s/%s", pod.Namespace, volume.PersistentVolumeClaim.ClaimName)
			pvWithPods, ok := claimedPVWithPods[pvcKey]
			if !ok {
				continue
			}

			pvWithPods.appendPodUnique(&pod)
		}
	}

	ret := make([]*persistentVolumeWithPods, 0, len(claimedPVWithPods))
	for _, v := range claimedPVWithPods {
		if len(v.pods) != 0 {
			ret = append(ret, v)
		}
	}

	return ret, nil
}

// getSecretReferenceByStorageClassOrPV
func getSecretReferenceByStorageClassOrPV(pv *corev1.PersistentVolume, scl *storagev1.StorageClassList) (*corev1.SecretReference, error) {
	if len(pv.Spec.StorageClassName) > 0 {
		for _, sc := range scl.Items {
			if pv.Namespace == sc.Namespace && pv.Spec.StorageClassName == sc.Name {
				return getNodeStageSecretReference(sc.Parameters, pv.GetName(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pv.Spec.ClaimRef.Name,
						Namespace: pv.Spec.ClaimRef.Namespace,
					},
				})
			}
		}
	}
	if pv.Spec.CSI != nil {
		return pv.Spec.CSI.NodeStageSecretRef, nil
	}
	return nil, nil
}

func IsCorruptedDir(dir string) bool {
	_, pathErr := mount.PathExists(dir)
	return pathErr != nil && mount.IsCorruptedMnt(pathErr)
}

// remountCorruptedVolumes try to remount all the volumes corrupted during csi-node restart,
// includes the GlobalMount per pv and BindMount per pod.
func (ns *nodeServer) remountCorruptedVolumes(ctx context.Context, nodeName string) {
	startTime := time.Now()

	pvWithPods, err := ns.getAttachedPVWithPodsOnNode(ctx, nodeName)
	if err != nil {
		klog.Warningf("get attached pv with pods info failed: %v\n", err)
		return
	}

	if len(pvWithPods) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(pvWithPods))
	for _, pvp := range pvWithPods {
		go func(p *persistentVolumeWithPods) {
			defer wg.Done()
			volumeId := p.Spec.CSI.VolumeHandle
			klog.Infof("remount ")
			// remount stagingTargetPath
			stagingTargetPath := filepath.Join(ns.kubeletRootDir, fmt.Sprintf("/plugins/kubernetes.io/csi/pv/%s/globalmount", p.Name))

			volOptions, err := NewVolOptionsFromVolIDOrStatic(volumeId, p.Spec.CSI.VolumeAttributes)
			if err != nil {
				klog.Warningf("New corrupted volume options %q error: %v\n", volumeId, err)
				return
			}
			volOptions.VolPath = stagingTargetPath

			fcfsMountOptions := &fcfs.MountOptionsSecrets{
				MountOptions: ns.mountOptions,
				Secrets:      p.credentials,
			}

			if err = ns.mounter.FcfsMount(ctx, volOptions, fcfsMountOptions); err != nil {
				klog.Warningf("remount corrupted volume %q to path %q failed: %v\n", p.Name, stagingTargetPath, err)
				return
			}
			klog.Infof("remount corrupted volume %q to global mount path %q succeed.", p.Name, stagingTargetPath)
			// TODO: mount options less?
			mountOptions := []string{"bind", "_netdev"}
			if p.Spec.CSI.ReadOnly {
				mountOptions = append(mountOptions, "ro")
			}
			// bind globalmount to pods
			for _, pod := range p.pods {
				podDir := filepath.Join(ns.kubeletRootDir, "/pods/", string(pod.UID))

				targetPath := filepath.Join(podDir, fmt.Sprintf("/volumes/kubernetes.io~csi/%s/mount", p.Name))

				if err := ns.mounter.Mount(stagingTargetPath, targetPath, p.Spec.CSI.FSType, mountOptions); err != nil {
					klog.Warningf("rebind corrupted volume %q to path %q failed: %v\n", p.Name, targetPath, err)
					continue
				}
				klog.Infof("rebind corrupted volume %q to pod mount path %q succeed.", p.Name, targetPath)

				// bind pod volume to subPath mount point
				for _, container := range pod.Spec.Containers {
					for i, volumeMount := range container.VolumeMounts {
						if volumeMount.SubPath == "" {
							continue
						}

						source := filepath.Join(targetPath, volumeMount.SubPath)

						// ref: https://github.com/kubernetes/kubernetes/blob/v1.22.0/pkg/volume/util/subpath/subpath_linux.go#L158
						subMountPath := filepath.Join(podDir, "volume-subpaths", p.Name, container.Name, strconv.Itoa(i))
						if err := ns.mounter.Mount(source, subMountPath, p.Spec.CSI.FSType, mountOptions); err != nil {
							klog.Warningf("rebind corrupted volume %q to sub mount path %q failed: %v\n", p.Name, subMountPath, err)
							continue
						}

						klog.Infof("rebind corrupted volume %q to sub mount path %q succeed.", p.Name, subMountPath)
					}
				}
			}
		}(pvp)
	}
	wg.Wait()

	klog.Infof("remount process finished cost %d ms", time.Since(startTime).Milliseconds())
}
