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

package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/codes"
	mount_fcfs_fused "vazmin.github.io/fastcfs-csi/pkg/fcfsfused-proxy/pb"
)

func TestServerMountAzureBlob(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    []string
		secrets map[string]string
		code    codes.Code
	}{
		{
			name:    "failed_mount",
			args:    []string{"--hello"},
			secrets: map[string]string{"hello": ""},
			code:    codes.InvalidArgument,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mountServer := NewMountServiceServer()
			req := mount_fcfs_fused.MountFcfsFusedRequest{
				MountArgs: tc.args,
				Secrets:   tc.secrets,
			}
			res, err := mountServer.MountFcfsFused(context.Background(), &req)
			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
			} else {
				require.Error(t, err)
				require.NotNil(t, res)
			}
		})
	}
}
