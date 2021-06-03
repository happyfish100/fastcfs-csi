package fcfs

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)

func fuseMount(ctx context.Context, volume *FcfsVolume, cr *common.Credentials) error {
	klog.V(5).Infof("fuse client mount volume %s", volume.VolID)
	if err := common.CreateDirIfNotExists(volume.VolPath); err != nil {
		return err
	}

	basePath := common.BuildBasePath(volume.VolName)
	if err := common.CreateDirIfNotExists(basePath); err != nil {
		return err
	}

	args := []string{
		"-u", cr.UserName,
		"-k", cr.KeyFile,
		"-b", basePath,
		"-n", volume.VolName,
		"-m", volume.VolPath,
		volume.getFuseClientConfigURL(), "restart",
	}

	output, err := common.ExecFuseCommand(ctx, args...)

	if err == nil {
		klog.V(5).Infof("[FastCFS] successfully fuse client mount")
		return nil
	}

	klog.Warningf("[FastCFS] failed to mount %s, output <= %s", volume.VolID, string(output))

	return err
}

func mountFcfsFusedWithProxy(ctx context.Context, volume *FcfsVolume, fcfsFusedProxyEndpoint string, fcfsFusedProxyConnTimout int, secrets map[string]string) (string, error) {
	klog.V(5).Infof("fuse client proxy mount volume %s", volume.VolID)
	if err := common.CreateDirIfNotExists(volume.VolPath); err != nil {
		return "", err
	}

	basePath := common.BuildBasePath(volume.VolName)
	args := []string{
		"-n", volume.VolName,
		"-m", volume.VolPath,
		volume.getFuseClientConfigURL(), "restart",
	}
	var resp *mount_fcfs_fused.MountFcfsFusedResponse
	var output string
	connectionTimout := time.Duration(fcfsFusedProxyConnTimout)
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimout*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fcfsFusedProxyEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err == nil {
		mountClient := NewMountClient(conn)
		mountReq := mount_fcfs_fused.MountFcfsFusedRequest{
			BasePath:  basePath,
			MountArgs: args,
			Secrets:   secrets,
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

func fuseUnmount(ctx context.Context, volume *FcfsVolume) error {
	klog.V(5).Infof("[FastCFS] fuse client unmount volume %s", volume.VolID)

	pid, err := common.GetPidFormBasePathByVolId(volume.VolName)
	if err != nil {
		return err
	}
	basePath := common.BuildBasePath(volume.VolName)
	p, err := os.FindProcess(pid)
	if err != nil {
		klog.Infof("[FastCFS] failed to find process %d: %v", pid, err)
	} else {
		if _, err = p.Wait(); err != nil {
			klog.Infof("[FastCFS] %d is not a child process: %v", pid, err)
		}
		if err := os.Remove(basePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
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
