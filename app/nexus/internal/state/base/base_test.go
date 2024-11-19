package base

// TODO: fixme.

//import (
//	"github.com/spiffe/spike/app/nexus/internal/state/store"
//	"testing"
//)
//
//func TestDeleteSecret(t *testing.T) {
//	originalKV := kv
//	defer func() {
//		kv = originalKV
//	}()
//
//	tests := []struct {
//		name    string
//		setup   func() *store.KV
//		wantErr error
//	}{
//		{
//			name: "success_case",
//			setup: func() *store.KV {
//				mockKV := store.NewKV()
//				mockKV.Put("test/path", map[string]string{"key": "value"})
//				return mockKV
//			},
//			wantErr: nil,
//		},
//		{
//			name: "delete_non_existent",
//			setup: func() *store.KV {
//				return store.NewKV()
//			},
//			wantErr: store.ErrSecretNotFound,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			kv = tt.setup()
//
//			err := DeleteSecret("test/path", nil)
//
//			if (err != nil && tt.wantErr == nil) ||
//				(err == nil && tt.wantErr != nil) ||
//				(err != nil && tt.wantErr != nil && err != tt.wantErr) {
//				t.Errorf("DeleteSecret() error = %v, wantErr %v", err, tt.wantErr)
//			}
//
//			if err == nil {
//				secret, getErr := kv.Get("test/path", 0)
//				if getErr != store.ErrSecretSoftDeleted {
//					t.Errorf("Expected secret to be soft deleted, got error: %v", getErr)
//				}
//				if secret != nil {
//					t.Errorf("Expected nil secret after deletion, got: %v", secret)
//				}
//			}
//		})
//	}
//}
//
//func TestDeleteSecretVersions(t *testing.T) {
//	originalKV := kv
//	defer func() {
//		kv = originalKV
//	}()
//
//	testKV := store.NewKV()
//	testKV.Put("test/path", map[string]string{"key": "v1"})
//	testKV.Put("test/path", map[string]string{"key": "v2"})
//	kv = testKV
//
//	tests := []struct {
//		name     string
//		versions []int
//		wantErr  error
//	}{
//		{
//			name:     "delete_specific_version",
//			versions: []int{1},
//			wantErr:  nil,
//		},
//		{
//			name:     "delete_current_version",
//			versions: []int{2},
//			wantErr:  nil,
//		},
//		{
//			name:     "delete_non_existent_version",
//			versions: []int{999},
//			wantErr:  nil,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			err := DeleteSecret("test/path", tt.versions)
//
//			if err != tt.wantErr {
//				t.Errorf("DeleteSecret() error = %v, wantErr %v", err, tt.wantErr)
//			}
//
//			if err == nil && len(tt.versions) > 0 {
//				for _, version := range tt.versions {
//					if version == 999 {
//						continue
//					}
//					secret, getErr := kv.Get("test/path", version)
//					if getErr != store.ErrSecretSoftDeleted {
//						t.Errorf("Expected version %d to be soft deleted, got error: %v", version, getErr)
//					}
//					if secret != nil {
//						t.Errorf("Expected nil secret for version %d after deletion, got: %v", version, secret)
//					}
//				}
//			}
//		})
//	}
//}
