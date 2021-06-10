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

package common

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
	TiB = GiB * 1024
)

const (
	CsiVolNamingPrefix = "csi-vol-"
)

const (
	PoolCMD              = "/usr/bin/fcfs_pool"
	PoolConfigFile       = "/fastcfs/auth/client.conf"
	FuseClientCMD        = "/usr/bin/fcfs_fused"
	FuseClientConfigFile = "/fastcfs/fcfs/fuse.conf"
)

const (
	DefaultDriverName  = "fcfs.csi.vazmin.github.io"
	DefaultCSIEndpoint = "unix://tmp/csi.sock"
)
