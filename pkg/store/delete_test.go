package store

import (
	"testing"
)

func TestKV_Delete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *KV
		path     string
		versions []int
		wantErr  error
	}{
		{
			name: "non_existent_path",
			setup: func() *KV {
				return &KV{
					data: make(map[string]*Secret),
				}
			},
			path:     "non/existent/path",
			versions: nil,
			wantErr:  ErrSecretNotFound,
		},
		{
			name: "delete_current_version_no_versions_specified",
			setup: func() *KV {
				kv := &KV{
					data: make(map[string]*Secret),
				}
				kv.data["test/path"] = &Secret{
					Metadata: Metadata{
						CurrentVersion: 1,
					},
					Versions: map[int]Version{
						1: {
							Data: map[string]string{
								"key": "test_value",
							},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: nil,
			wantErr:  nil,
		},
		{
			name: "delete_specific_versions",
			setup: func() *KV {
				kv := &KV{
					data: make(map[string]*Secret),
				}
				kv.data["test/path"] = &Secret{
					Metadata: Metadata{
						CurrentVersion: 2,
					},
					Versions: map[int]Version{
						1: {
							Data: map[string]string{
								"key": "value1",
							},
						},
						2: {
							Data: map[string]string{
								"key": "value2",
							},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: []int{1, 2},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := tt.setup()
			err := kv.Delete(tt.path, tt.versions)

			if err != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				secret, exists := kv.data[tt.path]
				if !exists {
					t.Errorf("Secret should still exist after deletion")
					return
				}

				if len(tt.versions) == 0 {
					cv := secret.Metadata.CurrentVersion
					if v, exists := secret.Versions[cv]; exists {
						if v.DeletedTime == nil {
							t.Errorf("Current version should be marked as deleted")
						}
					}
				} else {
					for _, version := range tt.versions {
						if version == 0 {
							version = secret.Metadata.CurrentVersion
						}
						if v, exists := secret.Versions[version]; exists {
							if v.DeletedTime == nil {
								t.Errorf("Version %d should be marked as deleted", version)
							}
						}
					}
				}
			}
		})
	}
}
