package driver

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	mountutils "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
	"strings"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
)

type Mounter interface {
	// Interface Implemented by NodeMounter.SafeFormatAndMount
	mountutils.Interface
	// Interface Implemented by NodeMounter.SafeFormatAndMount.Exec
	utilexec.Interface

	MakeDir(path string) error
	PathExists(path string) (bool, error)

	FcfsMount(ctx context.Context, volOptions *fcfs.VolumeOptions, mountOptions *fcfs.MountOptionsSecrets) error
}

type NodeMounter struct {
	mountutils.SafeFormatAndMount
	utilexec.Interface
}

func newNodeMounter() (Mounter, error) {
	safeMounter := &mountutils.SafeFormatAndMount{
		Interface: mountutils.New(""),
		Exec:      utilexec.New(),
	}
	return &NodeMounter{*safeMounter, safeMounter.Exec}, nil
}

func (n *NodeMounter) MakeDir(path string) error {
	return common.MakeDir(path)
}

func (n *NodeMounter) PathExists(path string) (bool, error) {
	return mountutils.PathExists(path)
}

func (n *NodeMounter) FcfsMount(ctx context.Context, volOptions *fcfs.VolumeOptions, mountOptions *fcfs.MountOptionsSecrets) error {

	if mountOptions.EnableFcfsFusedProxy {
		_, err := fcfs.MountFcfsFusedWithProxy(ctx, volOptions, mountOptions)
		return err
	} else {
		credentials, err := common.GetCredentialsForVolume(volOptions.PreProvisioned, mountOptions.Secrets)
		if err != nil {
			return err
		}
		defer credentials.DeleteCredentials()
		return fcfs.FuseMount(ctx, volOptions, credentials)
	}
}

func bindMount(ctx context.Context, from, to string, mntOptions []string) error {
	mntOptionSli := strings.Join(mntOptions, ",")

	output, err := common.ExecCommand(ctx, "mount", "-o", mntOptionSli, from, to)
	if err == nil {
		klog.V(5).Infof("successfully to mount bind.")
		return err
	}
	klog.Warningf("failed to mount bind, from %s, to %s output <= %s", from, to, string(output))
	return fmt.Errorf("failed to bind-mount %s to %s: %w", from, to, err)
}

func unmountVolume(ctx context.Context, mountPoint string) error {
	output, err := common.ExecCommand(ctx, "umount", mountPoint)
	if err != nil {
		klog.Warningf("unmount volume err, %s, %v", string(output), err)
		if strings.Contains(err.Error(), fmt.Sprintf("exit status 32: umount: %s: not mounted", mountPoint)) ||
			strings.Contains(err.Error(), "No such file or directory") {
			return nil
		}
		return err
	}

	return nil
}
