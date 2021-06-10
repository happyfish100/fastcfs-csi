package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"strconv"
	"vazmin.github.io/fastcfs-csi/pkg/common"
	"vazmin.github.io/fastcfs-csi/pkg/fcfs"
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


func newVolumeOptions(ctx context.Context, req *csi.CreateVolumeRequest, requestName string, cr *common.Credentials) (*fcfs.VolumeOptions, error) {
	parameters := req.GetParameters()
	clusterID := parameters["clusterID"]
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

	requiredBytes := common.RoundOffBytes(req.GetCapacityRange().GetRequiredBytes())

	return &fcfs.VolumeOptions{
		VolName:       cid.VolName,
		CapacityBytes: requiredBytes,
		VolID:         csiid,
		BaseConfigURL: url,
		ClusterID:     clusterID,
	}, nil
}


func NewVolOptionsFromVolID(volID string, cr *csi.CapacityRange) (*fcfs.VolumeOptions, error) {
	cid := &common.CSIIdentifier{}
	if err := cid.DecomposeCSIID(volID); err != nil {
		return nil, common.ErrInvalidVolID
	}

	url, err := common.ConfigURL(common.CsiConfigFile, cid.ClusterID)
	if err != nil {
		return nil, err
	}
	vol := &fcfs.VolumeOptions{
		VolName:       cid.VolName,
		VolID:         volID,
		ClusterID:     cid.ClusterID,
		BaseConfigURL: url,
	}
	if cr != nil {
		vol.CapacityBytes = cr.GetRequiredBytes()
	}
	return vol, nil
}


func NewVolOptionsFromStatic(volID string, options map[string]string) (*fcfs.VolumeOptions, error) {

	var (
		staticVol bool
		err       error
	)
	val, ok := options["static"]
	if !ok {
		return nil, common.ErrNonStaticVolume
	}

	if staticVol, err = strconv.ParseBool(val); err != nil {
		return nil, fmt.Errorf("failed to parse preProvisionedVolume: %w", err)
	}

	if !staticVol {
		return nil, common.ErrNonStaticVolume
	}

	clusterID, ok := options["clusterID"]
	if !ok {
		return nil, errors.New("clusterID must be set")
	}

	url, err := common.ConfigURL(common.CsiConfigFile, clusterID)
	if err != nil {
		return nil, err
	}

	vol := &fcfs.VolumeOptions{
		VolName:        volID,
		VolID:          volID,
		ClusterID:      clusterID,
		BaseConfigURL:  url,
		PreProvisioned: staticVol,
	}

	return vol, nil
}
