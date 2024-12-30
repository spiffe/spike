package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKV_Undelete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *KV
		path     string
		values   map[string]string
		versions []int
		wantErr  error
	}{
		{
			name: "undelete latest version if no versions specified",
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
							Data:        map[string]string{"key": "value"},
							Version:     1,
							DeletedTime: &time.Time{},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: []int{},
			wantErr:  nil,
		},
		{
			name: "undelete spesific versions",
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
							Data:        map[string]string{"key": "value1"},
							Version:     1,
							DeletedTime: &time.Time{},
						},
						2: {
							Data:        map[string]string{"key": "value2"},
							Version:     2,
							DeletedTime: &time.Time{},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: []int{1, 2},
			wantErr:  nil,
		},
		{
			name: "if secret does not exist",
			setup: func() *KV {
				return &KV{
					data:              make(map[string]*Secret),
					maxSecretVersions: 10,
				}
			},
			path:     "path/undelete/notExist",
			versions: []int{1},
			values:   map[string]string{"key": "value"},
			wantErr:  ErrSecretNotFound,
		},
		{
			name: "skip non-existent versions",
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
							Data:        map[string]string{"key": "value"},
							Version:     1,
							DeletedTime: &time.Time{},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: []int{1, 2},
			wantErr:  nil,
		},
		{
			name: "skip non-existent versions",
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
							Data:        map[string]string{"key": "value"},
							Version:     1,
							DeletedTime: &time.Time{},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			versions: []int{0},
			wantErr:  nil,
		},
		{
			name: "if version is 0 undelete current version",
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
							Data:        map[string]string{"key": "value"},
							Version:     1,
							DeletedTime: &time.Time{},
						},
					},
				}
				return kv
			},
			path:     "test/path",
			values:   map[string]string{"key": "value"},
			versions: []int{0},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := tt.setup()
			err := kv.Undelete(tt.path, tt.versions)
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				secret, exist := kv.data[tt.path]
				assert.True(t, exist)

				for _, version := range tt.versions {
					if version == 0 {
						version = secret.Metadata.CurrentVersion
					}
					if v, exist := secret.Versions[version]; exist {
						assert.True(t, exist)
						assert.Nil(t, v.DeletedTime)
					}

				}
			}
		})
	}
}
