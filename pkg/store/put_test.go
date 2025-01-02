package store

import (
	"testing"
)

func TestKV_Put(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *KV
		path     string
		values   map[string]string
		versions []int
		wantErr  error
	}{
		{
			setup: func() *KV {
				return &KV{
					data:              make(map[string]*Secret),
					maxSecretVersions: 10,
				}
			},
			name:     "it creates a new secret with initial metadata if the path doesn't exist",
			path:     "new/secret/path",
			versions: []int{1},
			values:   map[string]string{"key": "value"},
			wantErr:  nil,
		},
		{
			name: "it creates a new version with an incremented version number",
			setup: func() *KV {
				kv := &KV{data: make(map[string]*Secret), maxSecretVersions: 10}
				kv.Put("existing/secret/path", map[string]string{"key": "value1"})
				return kv
			},
			path:     "existing/secret/path",
			versions: []int{1, 2},
			wantErr:  nil,
		},
		{
			name: "it automatically prunes old versions when exceeding MaxVersions",
			setup: func() *KV {
				kv := &KV{data: make(map[string]*Secret), maxSecretVersions: 2}
				kv.Put("prune/old/versions", map[string]string{"key": "value1"})
				kv.Put("prune/old/versions", map[string]string{"key": "value2"})
				kv.Put("prune/old/versions", map[string]string{"key": "value3"})
				return kv
			},
			path:     "prune/old/versions",
			versions: []int{4, 3},
			values: map[string]string{
				"key": "value4",
			},
			wantErr: nil,
		},
		{
			name: "it updates timestamps for both creation and modification times",
			setup: func() *KV {
				kv := &KV{data: make(map[string]*Secret), maxSecretVersions: 10}
				kv.Put("update/timestamps", map[string]string{"key": "value1"})
				return kv
			},
			versions: []int{1, 2},
			path:     "update/timestamps",
			values:   map[string]string{"key": "value2"},
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := tt.setup()
			kv.Put(tt.path, tt.values)

			secret, exists := kv.data[tt.path]
			if !exists {
				t.Fatalf("expected secret to exist at path %q", tt.path)
			}

			if len(secret.Versions) != len(tt.versions) {
				t.Fatalf("expected %d versions, got %d", len(tt.versions), len(secret.Versions))
			}

			for _, version := range tt.versions {
				if _, exists := secret.Versions[version]; !exists {
					t.Fatalf("expected version %d to exist", version)
				}
			}

			if tt.wantErr != nil {
				t.Fatalf("unexpected error: %v", tt.wantErr)
			}
		})
	}
}
