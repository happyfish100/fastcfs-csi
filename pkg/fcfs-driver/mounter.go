package fcfs

import (
	"context"
	"fmt"
	"github.com/happyfish100/fastcfs-csi/pkg/common"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

func fuseMount(ctx context.Context, volume *FcfsVolume, cr *common.Credentials) error {
	klog.V(5).Infof("fuse client mount volume %s", volume.VolID)
	// os.MkdirAll("/opt/fastcfs/auth", os.ModePerm) // TODO: delete
	if err := common.CreateDirIfNotExists(volume.VolPath); err != nil {
		return err
	}

	basePath := common.BuildBasePath(volume.VolName)
	if err := common.CreateDirIfNotExists(basePath); err != nil {
		return err
	}
	//configFile, err := common.ConfigFile(common.CsiConfigFile, "1")
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

//
//func createFuseClientConf(volume *FcfsVolume) (string, error) {
//	input, err := ioutil.ReadFile(fuseClientConf)
//	if err != nil {
//		return "", err
//	}
//	output := bytes.Replace(input, []byte(mountpointVar), []byte(volume.VolPath), -1)
//	outputFinal := bytes.Replace(output, []byte(namespaceVar), []byte(volume.VolID), -1)
//	confPath := getConfPath(volume)
//	if err = ioutil.WriteFile(confPath, outputFinal, 0666); err != nil {
//		return "", err
//	}
//	return confPath, err
//}

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
