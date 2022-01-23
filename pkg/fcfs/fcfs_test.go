package fcfs

import (
	"context"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"testing"
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

func Test_cfs_VolumeExists(t *testing.T) {
	type args struct {
		ctx        context.Context
		baseURL    string
		volumeName string
		cr         *common.Credentials
	}
	baseURL := os.Getenv("FASTCFS_CONFIG_URL")
	keyFile := os.Getenv("FASTCFS_KEY_FILE")
	username := os.Getenv("FASTCFS_TEST_USERNAME")

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "volume already exists test cast",
			args: args{
				ctx:        context.TODO(),
				baseURL:    baseURL,
				volumeName: "foo",
				cr: &common.Credentials{
					KeyFile:  keyFile,
					UserName: username,
				},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "volume does not exist test cast",
			args: args{
				ctx:        context.TODO(),
				baseURL:    baseURL,
				volumeName: "bar",
				cr: &common.Credentials{
					KeyFile:  keyFile,
					UserName: username,
				},
			},
			want:    false,
			wantErr: false,
		}, {
			name: "bad credentials test cast",
			args: args{
				ctx:        context.TODO(),
				baseURL:    baseURL,
				volumeName: "bar",
				cr: &common.Credentials{
					KeyFile:  keyFile,
					UserName: rand.String(10),
				},
			},
			want:    true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cfs{}
			got, err := c.VolumeExists(tt.args.ctx, tt.args.baseURL, tt.args.volumeName, tt.args.cr)
			if (err != nil) != tt.wantErr {
				t.Errorf("VolumeExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VolumeExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
