package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_listNameSpace(t *testing.T) {
	tmpParent := t.TempDir()
	emptyDir := filepath.Join(tmpParent, "empty")
	os.Mkdir(emptyDir, 0755)

	populatedDir := filepath.Join(tmpParent, "populated")
	os.Mkdir(populatedDir, 0755)
	os.WriteFile(filepath.Join(populatedDir, "cmd.tsv"), []byte(""), 0644)
	os.WriteFile(filepath.Join(populatedDir, "dirs.tsv"), []byte(""), 0644)

	tests := []struct {
		name    string
		dir     string
		want    []string
		wantErr bool
	}{
		{
			name:    "invalid path",
			dir:     "/non/existent/path/at/all",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty",
			dir:     emptyDir,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "have some",
			dir:     populatedDir,
			want:    []string{"cmd", "dirs"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(tt.dir)

			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to create store: %v", err)
				}
				return
			}
			got, gotErr := s.ListNs()

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("listNameSpace() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listNameSpace() got = %v, want %v", got, tt.want)
			}
		})
	}
}
