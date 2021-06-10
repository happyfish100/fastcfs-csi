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
    "google.golang.org/grpc"
    mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)


type MountOptions struct {
    EnableFcfsFusedProxy     bool
    FcfsFusedEndpoint        string
    FcfsFusedProxyConnTimout int

}
type MountOptionsSecrets struct {
    *MountOptions
    Secrets   map[string]string
}



type MountClient struct {
    service mount_fcfs_fused.MountServiceClient
}

// NewMountClient returns a new mount client
func NewMountClient(cc *grpc.ClientConn) *MountClient {
    service := mount_fcfs_fused.NewMountServiceClient(cc)
    return &MountClient{service}
}

//func fuseUnmount(ctx context.Context, volume *FcfsVolume) error {
//	klog.V(5).Infof("[FastCFS] fuse client unmount volume %s", volume.VolID)
//
//	pid, err := common.GetPidFormBasePathByVolId(volume.VolName)
//	if err != nil {
//		return err
//	}
//	basePath := common.BuildBasePath(volume.VolName)
//	p, err := os.FindProcess(pid)
//	if err != nil {
//		klog.Infof("[FastCFS] failed to find process %d: %v", pid, err)
//	} else {
//		if _, err = p.Wait(); err != nil {
//			klog.Infof("[FastCFS] %d is not a child process: %v", pid, err)
//		}
//		if err := os.Remove(basePath); err != nil && !os.IsNotExist(err) {
//			return err
//		}
//	}
//	return nil
//}

