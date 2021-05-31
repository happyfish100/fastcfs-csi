/*
Copyright 2021 The Kubernetes Authors.

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

/*
Modifications Copyright 2021 vazmin.
Licensed under the Apache License, Version 2.0.
*/

package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"os/exec"
	"strings"
	"sync"
	"vazmin.github.io/fastcfs-csi/pkg/common"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)

var (
	mutex sync.Mutex
)

type MountServer struct {
	mount_fcfs_fused.UnimplementedMountServiceServer
}

// NewMountServiceServer returns a new Mountserver
func NewMountServiceServer() *MountServer {
	return &MountServer{}
}

// MountFcfsFused mounts an azure blob container to given location
func (server *MountServer) MountFcfsFused(ctx context.Context,
	req *mount_fcfs_fused.MountFcfsFusedRequest,
) (resp *mount_fcfs_fused.MountFcfsFusedResponse, err error) {
	mutex.Lock()
	defer mutex.Unlock()

	args := req.GetMountArgs()

	klog.V(2).Infof("received mount request: Mounting with args %v \n", args)

	cr, err := common.NewAdminCredentials(req.GetSecrets())
	if err != nil {
		klog.Errorf("failed to retrieve admin credentials: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer cr.DeleteCredentials()
	argsCr :=  []string{
		"-u", cr.UserName,
		"-k", cr.KeyFile,
	}
	var result mount_fcfs_fused.MountFcfsFusedResponse
	cmd := exec.Command("fcfs_fused", append(argsCr, strings.Split(args, " ")...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Error("fcfs_fused mount failed: with error:", err.Error())
	} else {
		klog.V(2).Infof("successfully mounted")
	}
	result.Output = string(output)
	klog.V(2).Infof("fcfs_fused output: %s\n", result.Output)
	return &result, err
}

func RunGRPCServer(
	mountServer mount_fcfs_fused.MountServiceServer,
	enableTLS bool,
	listener net.Listener,
) error {
	var serverOptions []grpc.ServerOption
	grpcServer := grpc.NewServer(serverOptions...)

	mount_fcfs_fused.RegisterMountServiceServer(grpcServer, mountServer)

	klog.V(2).Infof("Start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}
