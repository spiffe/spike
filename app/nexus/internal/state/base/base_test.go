package base

import (
	"errors"
	"testing"

	"github.com/spiffe/spike-sdk-go/kv"
)

func TestDeleteSecret(t *testing.T) {
	originalKV := secretStore
	defer func() {
		secretStore = originalKV
	}()

	tests := []struct {
		name    string
		setup   func() *kv.KV
		wantErr error
	}{
		{
			name: "success_case",
			setup: func() *kv.KV {
				mockKV := kv.New(kv.Config{MaxSecretVersions: 0})
				mockKV.Put("test/path", map[string]string{"key": "value"})
				return mockKV
			},
			wantErr: nil,
		},
		{
			name: "delete_non_existent",
			setup: func() *kv.KV {
				return kv.New(kv.Config{MaxSecretVersions: 0})
			},
			wantErr: kv.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretStore = tt.setup()

			err := DeleteSecret("test/path", nil)

			if (err != nil && tt.wantErr == nil) ||
				(err == nil && tt.wantErr != nil) ||
				(err != nil && tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("DeleteSecret() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				secret, getErr := secretStore.Get("test/path", 0)
				if !errors.Is(getErr, kv.ErrItemSoftDeleted) {
					t.Errorf("Expected secret to be soft deleted, got error: %v", getErr)
				}
				if secret != nil {
					t.Errorf("Expected nil secret after deletion, got: %v", secret)
				}
			}
		})
	}
}

func TestDeleteSecretVersions(t *testing.T) {
	originalKV := secretStore
	defer func() {
		secretStore = originalKV
	}()

	testKV := kv.New(kv.Config{MaxSecretVersions: 0})
	testKV.Put("test/path", map[string]string{"key": "v1"})
	testKV.Put("test/path", map[string]string{"key": "v2"})
	secretStore = testKV

	tests := []struct {
		name     string
		versions []int
		wantErr  error
	}{
		{
			name:     "delete_specific_version",
			versions: []int{1},
			wantErr:  nil,
		},
		{
			name:     "delete_current_version",
			versions: []int{2},
			wantErr:  nil,
		},
		{
			name:     "delete_non_existent_version",
			versions: []int{999},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteSecret("test/path", tt.versions)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteSecret() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && len(tt.versions) > 0 {
				for _, version := range tt.versions {
					if version == 999 {
						continue
					}
					secret, getErr := secretStore.Get("test/path", version)
					if !errors.Is(getErr, kv.ErrItemSoftDeleted) {
						t.Errorf("Expected version %d to be soft deleted, got error: %v", version, getErr)
					}
					if secret != nil {
						t.Errorf("Expected nil secret for version %d after deletion, got: %v", version, secret)
					}
				}
			}
		})
	}
}
