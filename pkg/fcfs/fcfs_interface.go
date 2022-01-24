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
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

type Cfs interface {
	CreateVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) (vol *Volume, err error)
	DeleteVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) (err error)
	ResizeVolume(ctx context.Context, volOptions *VolumeOptions, cr *common.Credentials) (newSize int64, err error)
	GetVolumeByID(ctx context.Context, volumeID string) (vol *Volume, err error)
	VolumeExists(ctx context.Context, configURL, volumeName string, cr *common.Credentials) (bool, error)
	MountVolume(ctx context.Context, volOptions *VolumeOptions, mountOptions *MountOptionsSecrets, cr *common.Credentials) error
}
