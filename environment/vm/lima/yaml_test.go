package lima

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/abiosoft/colima/config"
	"github.com/abiosoft/colima/util"
)

func Test_checkOverlappingMounts(t *testing.T) {
	type args struct {
		mounts []string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{args: args{mounts: []string{"/User", "/User/something"}}, wantErr: true},
		{args: args{mounts: []string{"/User/one", "/User/two"}}, wantErr: false},
		{args: args{mounts: []string{"/User/one", "/User/one_other"}}, wantErr: false},
		{args: args{mounts: []string{"/User/one_other", "/User/one"}}, wantErr: false},
		{args: args{mounts: []string{"/User/one", "/User/one/other"}}, wantErr: true},
		{args: args{mounts: []string{"/User/one/", "/User/one"}}, wantErr: true},
		{args: args{mounts: []string{"/User/one/", "/User/two", "User/one"}}, wantErr: true},
		{args: args{mounts: []string{"/home/a/b/c", "/home/b/c/a", "/home/c/a/b"}}, wantErr: false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			mounts := func(mounts []string) (mnts []config.Mount) {
				for _, m := range mounts {
					mnts = append(mnts, config.Mount{Location: m})
				}
				return
			}(tt.args.mounts)
			if err := checkOverlappingMounts(mounts); (err != nil) != tt.wantErr {
				t.Errorf("checkOverlappingMounts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_config_Mounts(t *testing.T) {
	tests := []struct {
		mounts        []string
		isDefault     bool
		includesCache bool
	}{
		{mounts: []string{"/User/user", "/tmp/another"}},
		{mounts: []string{"/User/another", "/User/something", "/User/else"}},
		{isDefault: true},
		{mounts: []string{util.HomeDir()}, includesCache: true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			mounts := func(mounts []string) (mnts []config.Mount) {
				for _, m := range mounts {
					mnts = append(mnts, config.Mount{Location: m})
				}
				return
			}(tt.mounts)
			conf, err := newConf(context.Background(), config.Config{Mounts: mounts})
			if err != nil {
				t.Error(err)
				return
			}

			expectedLocations := tt.mounts
			if tt.isDefault {
				expectedLocations = []string{"~", "/tmp/colima"}
			} else if !tt.includesCache {
				expectedLocations = append([]string{config.CacheDir()}, tt.mounts...)
			}

			sameMounts := func(expectedLocations []string, mounts []Mount) bool {
				sanitize := func(s string) string { return strings.TrimSuffix(s, "/") + "/" }
				for i, m := range mounts {
					if sanitize(m.Location) != sanitize(expectedLocations[i]) {
						return false
					}
				}
				return true
			}(expectedLocations, conf.Mounts)
			if !sameMounts {
				foundLocations := func() (locations []string) {
					for _, m := range conf.Mounts {
						locations = append(locations, m.Location)
					}
					return
				}()
				t.Errorf("got: %+v, want: %v", foundLocations, expectedLocations)
			}
		})
	}
}
