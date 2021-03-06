package common

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"io/ioutil"
	"k8s.io/klog/v2"
	utilpath "k8s.io/utils/path"
	"math"
	"os"
	"strconv"
)

func BuildBasePath(suffix string) string {
	return fmt.Sprintf("%s/%s", ClientBasePath, suffix)
}

func getPidFromBasePath(filepath string) (int, error) {
	pidFileBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidFileBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to parse FUSE daemon PID: %w", err)
	}
	return pid, nil
}

func CreateDirIfNotExists(dirPath string) error {
	exists, err := utilpath.Exists(utilpath.CheckFollowSymlink, dirPath)
	if err != nil {
		return err
	}
	if !exists {
		if err := os.Mkdir(dirPath, 0750); err != nil {
			if !os.IsExist(err) {
				klog.Errorf("failed to create path: %s with err: %v", dirPath, err)
				return err
			}
		}
	}
	return err
}

func GetPidFormBasePathByVolId(volId string) (int, error) {
	return getPidFromBasePath(fmt.Sprintf("%s/%s", BuildBasePath(volId), PidSuffixPath))
}

// checkDirExists checks directory exists or not.
func checkDirExists(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

// ConstructMountOptions returns only unique mount options in slice.
func ConstructMountOptions(mountOptions []string, volCap *csi.VolumeCapability) []string {
	if m := volCap.GetMount(); m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}
	return mountOptions
}

func RoundOffBytes(bytes int64) int64 {
	var num int64
	floatBytes := float64(bytes)
	// round off the value if its in decimal
	if floatBytes < GiB {
		// TODO: MiB
		num = GiB
	} else {
		num = int64(math.Ceil(floatBytes / GiB))
		num *= GiB
	}
	return num
}

// GiBToBytes converts GiB to Bytes
func GiBToBytes(volumeSizeGiB int64) int64 {
	return volumeSizeGiB * GiB
}

// RoundUpGiB rounds up the volume size in bytes upto multiplications of GiB
// in the unit of GiB
func RoundUpGiB(volumeSizeBytes int64) int64 {
	return roundUpSize(volumeSizeBytes, GiB)
}

// roundUpSize calculates how many allocation units are needed to accommodate a volume of given size.
// RoundUpSize(1500 * 1024*1024, 1024*1024*1024) returns '2'
// (2 GiB is the smallest allocatable volume that can hold 1500MiB)
func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp
}

func MakeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
