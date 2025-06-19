package common

import (
	"github.com/hashicorp/go-version"
	"reflect"
	"testing"
)

func TestMatchVersion(t *testing.T) {
	type args struct {
		v        string
		versions []*version.Version
	}
	tests := []struct {
		name    string
		args    args
		want    *version.Version
		wantErr bool
	}{
		{
			name: "latest",
			args: args{
				v: "latest",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1.2.5")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1.2",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1.2.5")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1.2",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.2.0")),
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1.2.5")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1.2.5")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1",
				versions: []*version.Version{
					version.Must(version.NewVersion("1")),
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.1")),
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
					version.Must(version.NewVersion("1.24.5-rc1")),
				},
			},
			want:    version.Must(version.NewVersion("1.2.5")),
			wantErr: false,
		},
		{
			name: "test-segments",
			args: args{
				v: "1",
				versions: []*version.Version{
					version.Must(version.NewVersion("1.1")),
					version.Must(version.NewVersion("1.2.3")),
					version.Must(version.NewVersion("1.2.4")),
					version.Must(version.NewVersion("1.2.5")),
					version.Must(version.NewVersion("1.2.5-rc1")),
					version.Must(version.NewVersion("1.24.5-rc1")),
					version.Must(version.NewVersion("1.24.5-rc2")),
					version.Must(version.NewVersion("1.24.5")),
				},
			},
			want:    version.Must(version.NewVersion("1.24.5")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchVersion(tt.args.v, tt.args.versions)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
