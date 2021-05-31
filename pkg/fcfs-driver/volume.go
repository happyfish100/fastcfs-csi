package fcfs

import (
	"context"
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	"math"
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

// createVolume
// fcfs_pool [-c config_filename=/etc/fastcfs/auth/client.conf]
//	[-u admin_username=admin]
//	[-k admin_secret_key_filename=/etc/fastcfs/auth/keys/${admin_username}.key]
//	[-d fastdir_access=rw]
//	[-s faststore_access=rw]
//	<operation> [username] [pool_name] [quota]
//
//	the operations and following parameters:
//	  create [pool_name] <quota> [--dryrun]
//	  quota <pool_name> <quota>
//	  delete | remove <pool_name>
//	  plist | pool-list [username] [pool_name]
//	  grant <username> <pool_name>
//	  cancel | withdraw <username> <pool_name>
//	  glist | grant-list | granted-list [username] [pool_name]
//
//	* the pool name can contain ${auto_id} for auto generated id when create pool, such as 'pool-${auto_id}'
//	  the pool name template configurated in the server side will be used when not specify the pool name
//	  you can set the initial value of auto increment id and the pool name template in server.conf of the server side
//
//	* the quota parameter is required for create and quota operations
//	  the default unit of quota is GB, unlimited for no limit
//
//	FastDIR and FastStore accesses are:
//	  r:  read only
//	  w:  write only
//	  rw: read and write

type FcfsVolume struct {
	//TopologyPools       *[]common.TopologyConstrainedPool
	TopologyRequirement *csi.TopologyRequirement
	Topology            map[string]string
	NamePrefix          string
	Size                int64
	VolID               string
	VolName             string
	VolPath             string
	ConfigURL           string
	ClusterID           string
}

func (fv *FcfsVolume) getPoolConfigURL() string {
	return fv.ConfigURL + common.PoolConfigFile
}

func (fv *FcfsVolume) getFuseClientConfigURL() string {
	return fv.ConfigURL + common.FuseClientConfigFile
}

// TODO: to obj func
func newFcfsVolume(ctx context.Context, req *csi.CreateVolumeRequest, requestName string, cr *common.Credentials) (*FcfsVolume, error) {

	parameters := req.GetParameters()
	clusterID := parameters["clusterID"]
	//configURL := parameters["configURL"]

	if len(clusterID) == 0 {
		return nil, errors.New("clusterID must be set")
	}

	cid := &common.CSIIdentifier{
		ClusterID: clusterID,
		UserName:  cr.UserName,
		VolName:   common.CsiVolNamingPrefix + requestName,
	}

	csiid, err := cid.ComposeCSIID()
	if err != nil {
		return nil, err
	}

	url, err := common.ConfigURL(common.CsiConfigFile, clusterID)
	if err != nil {
		return nil, err
	}
	// TODO topologies

	requiredBytes := common.RoundOffBytes(req.GetCapacityRange().GetRequiredBytes())

	return &FcfsVolume{
		VolName:   cid.VolName,
		Size:      requiredBytes,
		VolID:     csiid,
		ConfigURL: url,
		ClusterID: clusterID,
	}, nil
}

func newFcfsVolumeFromVolID(volID string, cr *csi.CapacityRange) (*FcfsVolume, error) {
	cid := &common.CSIIdentifier{}
	if err := cid.DecomposeCSIID(volID); err != nil {
		return nil, err
	}

	url, err := common.ConfigURL(common.CsiConfigFile, cid.ClusterID)
	if err != nil {
		return nil, err
	}
	vol := &FcfsVolume{
		VolName:   cid.VolName,
		VolID:     volID,
		ClusterID: cid.ClusterID,
		ConfigURL: url,
	}
	if cr != nil {
		vol.Size = common.RoundOffBytes(cr.GetRequiredBytes())
	}
	return vol, nil
}

func createVolume(ctx context.Context, volume *FcfsVolume, credentials *common.Credentials) error {

	args := []string{
		"-u", credentials.UserName,
		"-k", credentials.KeyFile,
		"-c", volume.getPoolConfigURL(),
		"create", volume.VolName,
		fmt.Sprintf("%dg", gibCeil(volume.Size)),
	}

	output, err := common.ExecPoolCommand(ctx, args...)
	if err != nil {
		klog.Errorf("[FastCFS] create volume %s", string(output))
		return err
	}
	klog.V(4).Infof("[FastCFS] successfully create FcfsVolume: %s", volume.VolID)
	return err
}

func volumeExists(ctx context.Context, volume *FcfsVolume, credentials *common.Credentials) bool {
	args := []string{
		"-u", credentials.UserName,
		"-k", credentials.KeyFile,
		"-c", volume.getPoolConfigURL(),
		"plist", credentials.UserName, volume.VolName,
	}
	_, err := common.ExecPoolCommand(ctx, args...)

	return err != nil
}

func resizeVolume(ctx context.Context, volume *FcfsVolume, credentials *common.Credentials) error {

	//os.MkdirAll("/opt/fastcfs/auth", os.ModePerm) // TODO: delete

	args := []string{
		"-u", credentials.UserName,
		"-k", credentials.KeyFile,
		"-c", volume.getPoolConfigURL(),
		"quota", volume.VolName,
		fmt.Sprintf("%dg", gibCeil(volume.Size)),
	}

	output, err := common.ExecPoolCommand(ctx, args...)

	if err != nil {
		klog.Warningf("[FastCFS] failed to resize FcfsVolume %s", string(output))
		return err
	}

	klog.V(4).Infof("[FastCFS] successfully resize FcfsVolume: %s", volume.VolID)
	return err
}

func gibCeil(size int64) int64 {
	return int64(math.Ceil(float64(size / common.GiB)))
}

func deleteVolume(ctx context.Context, vol *FcfsVolume, credentials *common.Credentials) error {
	klog.V(4).Infof("starting to delete FcfsVolume: %s", vol.VolID)

	args := []string{
		"-u", credentials.UserName,
		"-k", credentials.KeyFile,
		"-c", vol.getPoolConfigURL(),
		"delete", vol.VolName,
	}
	output, err := common.ExecPoolCommand(ctx, args...)

	if err != nil {
		klog.Warningf("[FastCFS] failed to delete FcfsVolume %s", string(output))
		return err
	}
	klog.V(4).Infof("[FastCFS] successfully deleted FcfsVolume: %s", vol.VolID)
	return nil
}
