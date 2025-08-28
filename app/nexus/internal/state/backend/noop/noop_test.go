//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package noop

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

func TestNoopStore_ImplementsBackendInterface(t *testing.T) {
	store := &NoopStore{}

	// Verify it implements Backend interface
	var _ backend.Backend = store
}

func TestNoopStore_Initialize(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	err := store.Initialize(ctx)

	if err != nil {
		t.Errorf("Initialize should return nil, got: %v", err)
	}
}

func TestNoopStore_InitializeWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := store.Initialize(ctx)

	if err != nil {
		t.Errorf("Initialize should return nil even with timeout, got: %v", err)
	}
}

func TestNoopStore_InitializeWithCancelledContext(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := store.Initialize(ctx)

	if err != nil {
		t.Errorf("Initialize should return nil even with cancelled context, got: %v", err)
	}
}

func TestNoopStore_Close(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	err := store.Close(ctx)

	if err != nil {
		t.Errorf("Close should return nil, got: %v", err)
	}
}

func TestNoopStore_CloseWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := store.Close(ctx)

	if err != nil {
		t.Errorf("Close should return nil even with timeout, got: %v", err)
	}
}

func TestNoopStore_CloseWithCancelledContext(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := store.Close(ctx)

	if err != nil {
		t.Errorf("Close should return nil even with cancelled context, got: %v", err)
	}
}

func TestNoopStore_LoadSecret(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	tests := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"simple path", "simple"},
		{"nested path", "app/database/credentials"},
		{"path with special chars", "app/service-1/api_key"},
		{"very long path", "very/long/path/that/goes/deep/into/the/hierarchy/with/many/segments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := store.LoadSecret(ctx, tt.path)

			if err != nil {
				t.Errorf("LoadSecret should return nil error, got: %v", err)
			}

			if secret != nil {
				t.Errorf("LoadSecret should return nil secret, got: %v", secret)
			}
		})
	}
}

func TestNoopStore_LoadSecretWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	secret, err := store.LoadSecret(ctx, "test/path")

	if err != nil {
		t.Errorf("LoadSecret should return nil error even with timeout, got: %v", err)
	}

	if secret != nil {
		t.Errorf("LoadSecret should return nil secret even with timeout, got: %v", secret)
	}
}

func TestNoopStore_LoadAllSecrets(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	secrets, err := store.LoadAllSecrets(ctx)

	if err != nil {
		t.Errorf("LoadAllSecrets should return nil error, got: %v", err)
	}

	if secrets != nil {
		t.Errorf("LoadAllSecrets should return nil map, got: %v", secrets)
	}
}

func TestNoopStore_LoadAllSecretsWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	secrets, err := store.LoadAllSecrets(ctx)

	if err != nil {
		t.Errorf("LoadAllSecrets should return nil error even with timeout, got: %v", err)
	}

	if secrets != nil {
		t.Errorf("LoadAllSecrets should return nil map even with timeout, got: %v", secrets)
	}
}

func TestNoopStore_StoreSecret(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	tests := []struct {
		name   string
		path   string
		secret kv.Value
	}{
		{
			name:   "empty path and secret",
			path:   "",
			secret: kv.Value{},
		},
		{
			name: "simple secret",
			path: "app/credentials",
			secret: kv.Value{
				Versions: map[int]kv.Version{
					1: {
						Data:    map[string]string{"username": "admin", "password": "secret"},
						Version: 1,
					},
				},
			},
		},
		{
			name: "multi-version secret",
			path: "app/api-key",
			secret: kv.Value{
				Versions: map[int]kv.Version{
					1: {
						Data:    map[string]string{"key": "old-key"},
						Version: 1,
					},
					2: {
						Data:    map[string]string{"key": "new-key"},
						Version: 2,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreSecret(ctx, tt.path, tt.secret)

			if err != nil {
				t.Errorf("StoreSecret should return nil, got: %v", err)
			}
		})
	}
}

func TestNoopStore_StoreSecretWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	secret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"test": "data"},
				Version: 1,
			},
		},
	}

	err := store.StoreSecret(ctx, "test/path", secret)

	if err != nil {
		t.Errorf("StoreSecret should return nil even with timeout, got: %v", err)
	}
}

func TestNoopStore_LoadPolicy(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"empty id", ""},
		{"simple id", "policy-1"},
		{"uuid-style id", "550e8400-e29b-41d4-a716-446655440000"},
		{"path-style id", "app/read-policy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := store.LoadPolicy(ctx, tt.id)

			if err != nil {
				t.Errorf("LoadPolicy should return nil error, got: %v", err)
			}

			if policy != nil {
				t.Errorf("LoadPolicy should return nil policy, got: %v", policy)
			}
		})
	}
}

func TestNoopStore_LoadPolicyWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	policy, err := store.LoadPolicy(ctx, "test-policy")

	if err != nil {
		t.Errorf("LoadPolicy should return nil error even with timeout, got: %v", err)
	}

	if policy != nil {
		t.Errorf("LoadPolicy should return nil policy even with timeout, got: %v", policy)
	}
}

func TestNoopStore_LoadAllPolicies(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	policies, err := store.LoadAllPolicies(ctx)

	if err != nil {
		t.Errorf("LoadAllPolicies should return nil error, got: %v", err)
	}

	if policies != nil {
		t.Errorf("LoadAllPolicies should return nil map, got: %v", policies)
	}
}

func TestNoopStore_LoadAllPoliciesWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	policies, err := store.LoadAllPolicies(ctx)

	if err != nil {
		t.Errorf("LoadAllPolicies should return nil error even with timeout, got: %v", err)
	}

	if policies != nil {
		t.Errorf("LoadAllPolicies should return nil map even with timeout, got: %v", policies)
	}
}

func TestNoopStore_StorePolicy(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	tests := []struct {
		name   string
		policy data.Policy
	}{
		{
			name:   "empty policy",
			policy: data.Policy{},
		},
		{
			name: "simple policy",
			policy: data.Policy{
				ID:              "read-policy",
				Name:            "Read Policy",
				SPIFFEIDPattern: "spiffe://example.org/reader/*",
				PathPattern:     "secrets/*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
		},
		{
			name: "multi-permission policy",
			policy: data.Policy{
				ID:              "admin-policy",
				Name:            "Admin Policy",
				SPIFFEIDPattern: "spiffe://example\\.org/admin/.*",
				PathPattern:     "admin/secret/.*",
				Permissions:     []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StorePolicy(ctx, tt.policy)

			if err != nil {
				t.Errorf("StorePolicy should return nil, got: %v", err)
			}
		})
	}
}

func TestNoopStore_StorePolicyWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	policy := data.Policy{
		ID:              "test-policy",
		Name:            "Test Policy",
		SPIFFEIDPattern: "spiffe://example.org/test/*",
		PathPattern:     "test/*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}

	err := store.StorePolicy(ctx, policy)

	if err != nil {
		t.Errorf("StorePolicy should return nil even with timeout, got: %v", err)
	}
}

func TestNoopStore_DeletePolicy(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"empty id", ""},
		{"simple id", "policy-to-delete"},
		{"non-existent id", "does-not-exist"},
		{"uuid-style id", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeletePolicy(ctx, tt.id)

			if err != nil {
				t.Errorf("DeletePolicy should return nil, got: %v", err)
			}
		})
	}
}

func TestNoopStore_DeletePolicyWithTimeout(t *testing.T) {
	store := &NoopStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := store.DeletePolicy(ctx, "test-policy")

	if err != nil {
		t.Errorf("DeletePolicy should return nil even with timeout, got: %v", err)
	}
}

func TestNoopStore_GetCipher(t *testing.T) {
	store := &NoopStore{}

	cipher := store.GetCipher()

	if cipher != nil {
		t.Errorf("GetCipher should return nil, got: %v", cipher)
	}
}

func TestNoopStore_ConcurrentOperations(t *testing.T) {
	store := &NoopStore{}
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent access to all methods
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Test all methods concurrently
			_ = store.Initialize(ctx)
			_ = store.Close(ctx)
			_, _ = store.LoadSecret(ctx, "concurrent/secret")
			_, _ = store.LoadAllSecrets(ctx)
			_, _ = store.LoadPolicy(ctx, "concurrent-policy")
			_, _ = store.LoadAllPolicies(ctx)
			_ = store.DeletePolicy(ctx, "concurrent-delete")
			store.GetCipher()

			// Store operations
			secret := kv.Value{
				Versions: map[int]kv.Version{
					1: {
						Data:    map[string]string{"concurrent": "test"},
						Version: 1,
					},
				},
			}
			_ = store.StoreSecret(ctx, "concurrent/test", secret)

			policy := data.Policy{
				ID:              "concurrent-policy",
				Name:            "Concurrent Policy",
				SPIFFEIDPattern: "spiffe://example.org/concurrent/*",
				PathPattern:     "concurrent/*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}
			_ = store.StorePolicy(ctx, policy)
		}(i)
	}

	wg.Wait()

	// All operations should complete without error
	// No assertions needed since all methods are no-ops
	t.Log("All concurrent operations completed successfully")
}

func TestNoopStore_MultipleInstances(t *testing.T) {
	ctx := context.Background()

	// Create multiple instances
	store1 := &NoopStore{}
	store2 := &NoopStore{}
	store3 := &NoopStore{}

	stores := []*NoopStore{store1, store2, store3}

	// Test that each instance behaves consistently
	for i, store := range stores {
		t.Run(fmt.Sprintf("instance_%d", i), func(t *testing.T) {
			// Initialize
			err := store.Initialize(ctx)
			if err != nil {
				t.Errorf("Initialize failed on instance %d: %v", i, err)
			}

			// Test secret operations
			secret, err := store.LoadSecret(ctx, "test/path")
			if err != nil || secret != nil {
				t.Errorf("LoadSecret failed on instance %d: err=%v, secret=%v", i, err, secret)
			}

			secrets, err := store.LoadAllSecrets(ctx)
			if err != nil || secrets != nil {
				t.Errorf("LoadAllSecrets failed on instance %d: err=%v, secrets=%v", i, err, secrets)
			}

			testSecret := kv.Value{
				Versions: map[int]kv.Version{
					1: {
						Data:    map[string]string{"test": "value"},
						Version: 1,
					},
				},
			}
			err = store.StoreSecret(ctx, "test/path", testSecret)
			if err != nil {
				t.Errorf("StoreSecret failed on instance %d: %v", i, err)
			}

			// Test policy operations
			policy, err := store.LoadPolicy(ctx, "test-policy")
			if err != nil || policy != nil {
				t.Errorf("LoadPolicy failed on instance %d: err=%v, policy=%v", i, err, policy)
			}

			policies, err := store.LoadAllPolicies(ctx)
			if err != nil || policies != nil {
				t.Errorf("LoadAllPolicies failed on instance %d: err=%v, policies=%v", i, err, policies)
			}

			testPolicy := data.Policy{
				ID:              "test-policy",
				Name:            "Test Policy",
				SPIFFEIDPattern: "spiffe://example.org/test/*",
				PathPattern:     "test/*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}
			err = store.StorePolicy(ctx, testPolicy)
			if err != nil {
				t.Errorf("StorePolicy failed on instance %d: %v", i, err)
			}

			err = store.DeletePolicy(ctx, "test-policy")
			if err != nil {
				t.Errorf("DeletePolicy failed on instance %d: %v", i, err)
			}

			// Test cipher
			cipher := store.GetCipher()
			if cipher != nil {
				t.Errorf("GetCipher should return nil on instance %d, got: %v", i, cipher)
			}

			// Close
			err = store.Close(ctx)
			if err != nil {
				t.Errorf("Close failed on instance %d: %v", i, err)
			}
		})
	}
}

func TestNoopStore_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	store := &NoopStore{}
	ctx := context.Background()

	// Perform many operations rapidly
	numOperations := 10000

	for i := 0; i < numOperations; i++ {
		// Mix of all operations
		switch i % 8 {
		case 0:
			_ = store.Initialize(ctx)
		case 1:
			_, _ = store.LoadSecret(ctx, "stress/test")
		case 2:
			_, _ = store.LoadAllSecrets(ctx)
		case 3:
			secret := kv.Value{
				Versions: map[int]kv.Version{
					1: {Data: map[string]string{"stress": "test"}, Version: 1},
				},
			}
			_ = store.StoreSecret(ctx, "stress/test", secret)
		case 4:
			_, _ = store.LoadPolicy(ctx, "stress-policy")
		case 5:
			_, _ = store.LoadAllPolicies(ctx)
		case 6:
			policy := data.Policy{
				ID:              "stress-policy",
				Name:            "Stress Policy",
				SPIFFEIDPattern: "spiffe://example.org/stress/*",
				PathPattern:     "stress/*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}
			_ = store.StorePolicy(ctx, policy)
		case 7:
			_ = store.DeletePolicy(ctx, "stress-policy")
		}
	}

	// Final operations
	store.GetCipher()
	_ = store.Close(ctx)

	t.Logf("Completed %d stress test operations successfully", numOperations)
}
