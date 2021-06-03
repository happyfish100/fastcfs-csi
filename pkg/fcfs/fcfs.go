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
    "fmt"
    "github.com/container-storage-interface/spec/lib/go/csi"
    "google.golang.org/grpc"
    "k8s.io/klog/v2"
    "time"
    "vazmin.github.io/fastcfs-csi/pkg/common"
    mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)

type cfs struct {
}

type Volume struct {
    Labels        map[string]string
    CapacityBytes int64
    VolumeId      string
}

type VolumeOptions struct {
    CapacityBytes       int64
    TopologyRequirement *csi.TopologyRequirement
    Topology            map[string]string
    VolID               string
    VolName             string
    VolPath             string
    BaseConfigURL       string
    ClusterID           string
}


func (vo *VolumeOptions) getPoolConfigURL() string {
    return vo.BaseConfigURL + common.PoolConfigFile
}

func (vo *VolumeOptions) getFuseClientConfigURL() string {
    return vo.BaseConfigURL + common.FuseClientConfigFile
}

var _ Cfs = &cfs{}

func NewCFS() (Cfs, error) {
    return &cfs{}, nil
}

func (c *cfs) CreateVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) (*Volume, error) {
    args := []string{
        "-u", cr.UserName,
        "-k", cr.KeyFile,
        "-c", volOptions.getPoolConfigURL(),
        "create", volOptions.VolName,
        fmt.Sprintf("%dg", common.RoundUpGiB(volOptions.CapacityBytes)),
    }

    output, err := common.ExecPoolCommand(ctx, args...)
    if err != nil {
        klog.Errorf("[FastCFS] create volume %s", string(output))
        return nil, err
    }
    klog.V(4).Infof("[FastCFS] successfully create FcfsVolume: %s", volOptions.VolID)

    return &Volume{
        VolumeId: volOptions.VolID,
        CapacityBytes: volOptions.CapacityBytes,
        Labels: volOptions.Topology,
    }, nil
}

func (c *cfs) VolumeExists(ctx context.Context, baseURL, volumeName string, cr *common.Credentials) (bool, error) {
    args := []string{
        "-u", cr.UserName,
        "-k", cr.KeyFile,
        "-c", baseURL + common.PoolConfigFile,
        "plist", cr.UserName, volumeName,
    }
    _, err := common.ExecPoolCommand(ctx, args...)
    return err != nil, nil
}


func (c *cfs) DeleteVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) ( err error) {
    args := []string{
        "-u", cr.UserName,
        "-k", cr.KeyFile,
        "-c", volOptions.getPoolConfigURL(),
        "delete", volOptions.VolName,
    }
    output, err := common.ExecPoolCommand(ctx, args...)

    if err != nil {
        klog.Warningf("[FastCFS] failed to delete FcfsVolume %s", string(output))
        return
    }
    klog.V(4).Infof("[FastCFS] successfully deleted FcfsVolume: %s", volOptions.VolID)
    return
}

func (c *cfs) ResizeVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) (int64, error) {
    newSize := common.RoundUpGiB(volOptions.CapacityBytes)
    args := []string{
        "-u", cr.UserName,
        "-k", cr.KeyFile,
        "-c", volOptions.getPoolConfigURL(),
        "quota", volOptions.VolName,
        fmt.Sprintf("%dg", common.RoundUpGiB(volOptions.CapacityBytes)),
    }

    output, err := common.ExecPoolCommand(ctx, args...)

    if err != nil {
        klog.Warningf("[FastCFS] failed to resize FcfsVolume %s", string(output))
        return 0, err
    }

    klog.V(4).Infof("[FastCFS] successfully resize FcfsVolume: %s", volOptions.VolID)
    return newSize, nil
}

func (c *cfs) MountVolume(ctx context.Context, volOptions *VolumeOptions, mountOptions *MountOptionsSecrets, cr *common.Credentials) error {
    if mountOptions.EnableFcfsFusedProxy {
        _, err := MountFcfsFusedWithProxy(ctx, volOptions, mountOptions)
        return err
    } else {
        return FuseMount(ctx, volOptions, cr)
    }
}

func (c *cfs) GetVolumeByID(ctx context.Context,  volumeID string) (disk *Volume, err error) {
    panic("implement me")
}

func FuseMount(ctx context.Context, volumeOptions *VolumeOptions,  cr *common.Credentials) error {
    klog.V(5).Infof("fuse client mount volume %s", volumeOptions.VolID)
    if err := common.CreateDirIfNotExists(volumeOptions.VolPath); err != nil {
        return err
    }

    basePath := common.BuildBasePath(volumeOptions.VolName)
    if err := common.CreateDirIfNotExists(basePath); err != nil {
        return err
    }

    args := []string{
        "-u", cr.UserName,
        "-k", cr.KeyFile,
        "-b", basePath,
        "-n", volumeOptions.VolName,
        "-m", volumeOptions.VolPath,
        volumeOptions.BaseConfigURL + common.FuseClientConfigFile, "restart",
    }

    output, err := common.ExecFuseCommand(ctx, args...)

    if err == nil {
        klog.V(5).Infof("[FastCFS] successfully fuse client mount")
        return nil
    }

    klog.Warningf("[FastCFS] failed to mount %s, output <= %s", volumeOptions.VolID, string(output))

    return err
}

func MountFcfsFusedWithProxy(ctx context.Context, volumeOptions *VolumeOptions, mountOption *MountOptionsSecrets) (string, error) {
    klog.V(5).Infof("fuse client proxy mount volume %s", volumeOptions.VolID)
    if err := common.CreateDirIfNotExists(volumeOptions.VolPath); err != nil {
        return "", err
    }

    basePath := common.BuildBasePath(volumeOptions.VolName)
    args := []string{
        "-n", volumeOptions.VolName,
        "-m", volumeOptions.VolPath,
        volumeOptions.BaseConfigURL + common.FuseClientConfigFile, "restart",
    }
    var resp *mount_fcfs_fused.MountFcfsFusedResponse
    var output string
    connectionTimout := time.Duration(mountOption.FcfsFusedProxyConnTimout)
    ctx, cancel := context.WithTimeout(context.Background(), connectionTimout*time.Second)
    defer cancel()
    conn, err := grpc.DialContext(ctx, mountOption.FcfsFusedEndpoint, grpc.WithInsecure(), grpc.WithBlock())
    if err == nil {
        mountClient := NewMountClient(conn)
        mountReq := mount_fcfs_fused.MountFcfsFusedRequest{
            BasePath:  basePath,
            MountArgs: args,
            Secrets:   mountOption.Secrets,
        }
        klog.V(2).Infof("calling fcfsfused Proxy: MountFcfsFused function")
        resp, err = mountClient.service.MountFcfsFused(context.TODO(), &mountReq)
        if err != nil {
            klog.Error("GRPC call returned with an error:", err)
        }
        output = resp.GetOutput()
    }
    return output, err
}

