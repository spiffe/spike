//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

func TestNewInMemoryStore(t *testing.T) {
	testCipher := createTestCipher(t)
	maxVersions := 10

	store := NewInMemoryStore(testCipher, maxVersions)

	// Verify store was created properly
	if store == nil {
		t.Fatal("Expected non-nil Store")
		return
	}

	if store.secretStore == nil {
		t.Fatal("Expected non-nil secretStore")
	}

	if store.policies == nil {
		t.Fatal("Expected non-nil policies map")
	}

	if store.cipher != testCipher {
		t.Fatal("Expected cipher to be set correctly")
	}

	// Verify it implements Backend interface
	var _ backend.Backend = store
}

func TestInMemoryStore_Initialize(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)

	ctx := context.Background()
	initErr := store.Initialize(ctx)

	if initErr != nil {
		t.Errorf("Initialize should not return error: %v", initErr)
	}
}

func TestInMemoryStore_Close(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)

	ctx := context.Background()
	closeErr := store.Close(ctx)

	if closeErr != nil {
		t.Errorf("Close should not return error: %v", closeErr)
	}
}

func TestInMemoryStore_GetCipher(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)

	retrievedCipher := store.GetCipher()

	if retrievedCipher != testCipher {
		t.Error("GetCipher should return the same cipher instance")
	}

	if retrievedCipher == nil {
		t.Error("GetCipher should not return nil")
	}
}

func TestInMemoryStore_StoreAndLoadSecret(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Create a test secret
	secret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:        map[string]string{"username": "admin", "password": "secret123"},
				Version:     1,
				CreatedTime: time.Now(),
			},
		},
	}

	path := "test/secret/path"

	// Store the secret
	storeErr := store.StoreSecret(ctx, path, secret)
	if storeErr != nil {
		t.Errorf("StoreSecret failed: %v", storeErr)
	}

	// Load the secret
	loadedSecret, loadErr := store.LoadSecret(ctx, path)
	if loadErr != nil {
		t.Errorf("LoadSecret failed: %v", loadErr)
	}

	if loadedSecret == nil {
		t.Fatal("Expected non-nil loaded secret")
		return
	}

	// Verify the loaded secret matches the stored one
	if len(loadedSecret.Versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(loadedSecret.Versions))
	}

	version1, exists := loadedSecret.Versions[1]
	if !exists {
		t.Error("Expected version 1 to exist")
	}

	if version1.Data["username"] != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", version1.Data["username"])
	}

	if version1.Data["password"] != "secret123" {
		t.Errorf("Expected password 'secret123', got '%s'", version1.Data["password"])
	}
}

func TestInMemoryStore_LoadNonExistentSecret(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Try to load a secret that doesn't exist
	loadedSecret, loadErr := store.LoadSecret(ctx, "nonexistent/path")

	// Should return ErrEntityNotFound for a non-existent secret
	if loadErr == nil {
		t.Error("Expected ErrEntityNotFound for non-existent secret")
	}

	if loadErr != nil && !loadErr.Is(sdkErrors.ErrEntityNotFound) {
		t.Errorf("Expected ErrEntityNotFound, got: %v", loadErr)
	}

	// Should return nil secret
	if loadedSecret != nil {
		t.Errorf("Expected nil secret for non-existent path, got: %v", loadedSecret)
	}
}

func TestInMemoryStore_LoadAllSecrets(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Store multiple secrets
	secrets := map[string]kv.Value{
		"app1/database": {
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"host": "db1.example.com", "port": "5432"},
					Version: 1,
				},
			},
		},
		"app2/api_key": {
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "api-key-123"},
					Version: 1,
				},
			},
		},
		"shared/config": {
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"env": "production"},
					Version: 1,
				},
			},
		},
	}

	// Store all secrets
	for path, secret := range secrets {
		storeErr := store.StoreSecret(ctx, path, secret)
		if storeErr != nil {
			t.Errorf("Failed to store secret at %s: %v", path, storeErr)
		}
	}

	// Load all secrets
	allSecrets, loadErr := store.LoadAllSecrets(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllSecrets failed: %v", loadErr)
	}

	if len(allSecrets) != len(secrets) {
		t.Errorf("Expected %d secrets, got %d", len(secrets), len(allSecrets))
	}

	// Verify each secret was loaded correctly
	for path, expectedSecret := range secrets {
		loadedSecret, exists := allSecrets[path]
		if !exists {
			t.Errorf("Expected secret at path %s to exist", path)
			continue
		}

		if len(loadedSecret.Versions) != len(expectedSecret.Versions) {
			t.Errorf("Version count mismatch for %s: expected %d, got %d",
				path, len(expectedSecret.Versions), len(loadedSecret.Versions))
		}

		// Check version 1 data
		expectedVersion := expectedSecret.Versions[1]
		loadedVersion := loadedSecret.Versions[1]

		if !reflect.DeepEqual(expectedVersion.Data, loadedVersion.Data) {
			t.Errorf("Data mismatch for %s: expected %v, got %v",
				path, expectedVersion.Data, loadedVersion.Data)
		}
	}
}

func TestInMemoryStore_LoadAllSecretsEmpty(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Load all secrets from the empty store
	allSecrets, loadErr := store.LoadAllSecrets(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllSecrets failed: %v", loadErr)
	}

	if len(allSecrets) != 0 {
		t.Errorf("Expected empty map, got %d secrets", len(allSecrets))
	}
}

func TestInMemoryStore_StoreAndLoadPolicy(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Create a test policy
	policy := data.Policy{
		Name:            "test-policy-1",
		SPIFFEIDPattern: "^spiffe://example\\.org/app/.*$",
		PathPattern:     "^app/secrets/.*$",
		Permissions:     []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
	}

	// Store the policy
	storeErr := store.StorePolicy(ctx, policy)
	if storeErr != nil {
		t.Errorf("StorePolicy failed: %v", storeErr)
	}

	// Load the policy
	loadedPolicy, loadErr := store.LoadPolicy(ctx, policy.Name)
	if loadErr != nil {
		t.Errorf("LoadPolicy failed: %v", loadErr)
	}

	if loadedPolicy == nil {
		t.Fatal("Expected non-nil loaded policy")
		return
	}

	// Verify the loaded policy matches the stored one
	if loadedPolicy.Name != policy.Name {
		t.Errorf("Expected Name '%s', got '%s'", policy.Name, loadedPolicy.Name)
	}

	if loadedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
		t.Errorf("Expected SPIFFEIDPattern '%s', got '%s'",
			policy.SPIFFEIDPattern, loadedPolicy.SPIFFEIDPattern)
	}

	if loadedPolicy.PathPattern != policy.PathPattern {
		t.Errorf("Expected PathPattern '%s', got '%s'",
			policy.PathPattern, loadedPolicy.PathPattern)
	}

	if !reflect.DeepEqual(loadedPolicy.Permissions, policy.Permissions) {
		t.Errorf("Expected Permissions %v, got %v",
			policy.Permissions, loadedPolicy.Permissions)
	}
}

func TestInMemoryStore_StorePolicyEmptyName(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Create a policy with empty Name
	policy := data.Policy{
		Name:            "", // Empty Name
		SPIFFEIDPattern: "^spiffe://example\\.org/app/.*$",
		PathPattern:     "^app/secrets/.*$",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}

	// Store the policy - should fail with ErrEntityInvalid
	storeErr := store.StorePolicy(ctx, policy)
	if storeErr == nil {
		t.Error("Expected error when storing policy with empty name")
	}

	if storeErr != nil && !storeErr.Is(sdkErrors.ErrEntityInvalid) {
		t.Errorf("Expected ErrEntityInvalid, got: %v", storeErr)
	}
}

func TestInMemoryStore_LoadNonExistentPolicy(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Try to load a policy that doesn't exist
	loadedPolicy, loadErr := store.LoadPolicy(ctx, "nonexistent-policy")

	// Should return ErrEntityNotFound for a non-existent policy
	if loadErr == nil {
		t.Error("Expected ErrEntityNotFound for non-existent policy")
	}

	if loadErr != nil && !loadErr.Is(sdkErrors.ErrEntityNotFound) {
		t.Errorf("Expected ErrEntityNotFound, got: %v", loadErr)
	}

	if loadedPolicy != nil {
		t.Error("Expected nil policy for non-existent ID")
	}
}

func TestInMemoryStore_LoadAllPolicies(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Store multiple policies
	policies := map[string]data.Policy{
		"policy-1": {
			Name:            "policy-1",
			SPIFFEIDPattern: "^spiffe://example\\.org/reader/.*$",
			PathPattern:     "^read/.*$",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		},
		"policy-2": {
			Name:            "policy-2",
			SPIFFEIDPattern: "^spiffe://example\\.org/writer/.*$",
			PathPattern:     "^write/.*$",
			Permissions:     []data.PolicyPermission{data.PermissionWrite},
		},
		"policy-3": {
			Name:            "policy-3",
			SPIFFEIDPattern: "^spiffe://example\\.org/admin/.*$",
			PathPattern:     "^secrets/.*$",
			Permissions: []data.PolicyPermission{data.PermissionRead,
				data.PermissionWrite, data.PermissionList},
		},
	}

	// Store all policies
	for _, policy := range policies {
		storeErr := store.StorePolicy(ctx, policy)
		if storeErr != nil {
			t.Errorf("Failed to store policy %s: %v", policy.Name, storeErr)
		}
	}

	// Load all policies
	allPolicies, loadErr := store.LoadAllPolicies(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllPolicies failed: %v", loadErr)
	}

	if len(allPolicies) != len(policies) {
		t.Errorf("Expected %d policies, got %d", len(policies), len(allPolicies))
	}

	// Verify each policy was loaded correctly
	for name, expectedPolicy := range policies {
		loadedPolicy, exists := allPolicies[name]
		if !exists {
			t.Errorf("Expected policy with name %s to exist", name)
			continue
		}

		if loadedPolicy.Name != expectedPolicy.Name {
			t.Errorf("Name mismatch for %s: expected %s, got %s",
				name, expectedPolicy.Name, loadedPolicy.Name)
		}

		if !reflect.DeepEqual(loadedPolicy.Permissions, expectedPolicy.Permissions) {
			t.Errorf("Permissions mismatch for %s: expected %v, got %v",
				name, expectedPolicy.Permissions, loadedPolicy.Permissions)
		}
	}
}

func TestInMemoryStore_LoadAllPoliciesEmpty(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Load all policies from the empty store
	allPolicies, loadErr := store.LoadAllPolicies(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllPolicies failed: %v", loadErr)
	}

	if len(allPolicies) != 0 {
		t.Errorf("Expected empty map, got %d policies", len(allPolicies))
	}
}

func TestInMemoryStore_DeletePolicy(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Create and store test policy
	policy := data.Policy{
		Name:            "deletable-policy",
		SPIFFEIDPattern: "^spiffe://example\\.org/temp/.*$",
		PathPattern:     "^secrets/temp/.*$",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}

	storeErr := store.StorePolicy(ctx, policy)
	if storeErr != nil {
		t.Fatalf("Failed to store test policy: %v", storeErr)
	}

	// Verify policy exists
	loadedPolicy, loadErr := store.LoadPolicy(ctx, policy.Name)
	if loadErr != nil || loadedPolicy == nil {
		t.Fatal("Policy should exist before deletion")
	}

	// Delete the policy
	deleteErr := store.DeletePolicy(ctx, policy.Name)
	if deleteErr != nil {
		t.Errorf("DeletePolicy failed: %v", deleteErr)
	}

	// Verify policy no longer exists (LoadPolicy returns ErrEntityNotFound)
	deletedPolicy, loadAfterDeleteErr := store.LoadPolicy(ctx, policy.Name)
	if loadAfterDeleteErr == nil || !loadAfterDeleteErr.Is(sdkErrors.ErrEntityNotFound) {
		t.Errorf("Expected ErrEntityNotFound after deletion, got: %v", loadAfterDeleteErr)
	}

	if deletedPolicy != nil {
		t.Error("Policy should not exist after deletion")
	}
}

func TestInMemoryStore_DeleteNonExistentPolicy(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	// Delete a policy that doesn't exist
	deleteErr := store.DeletePolicy(ctx, "nonexistent-policy")

	// Should not return error
	if deleteErr != nil {
		t.Errorf("DeletePolicy should not return error for non-existent policy: %v", deleteErr)
	}
}

func TestInMemoryStore_ConcurrentSecretOperations(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	secretsPerGoroutine := 5

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < secretsPerGoroutine; j++ {
				path := fmt.Sprintf("concurrent/secret-%d-%d", goroutineID, j)
				secret := kv.Value{
					Versions: map[int]kv.Version{
						1: {
							Data:    map[string]string{"data": fmt.Sprintf("value-%d-%d", goroutineID, j)},
							Version: 1,
						},
					},
				}

				err := store.StoreSecret(ctx, path, secret)
				if err != nil {
					t.Errorf("Concurrent StoreSecret failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all secrets were stored
	allSecrets, loadErr := store.LoadAllSecrets(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllSecrets after concurrent writes failed: %v", loadErr)
	}

	expectedCount := numGoroutines * secretsPerGoroutine
	if len(allSecrets) != expectedCount {
		t.Errorf("Expected %d secrets after concurrent writes, got %d",
			expectedCount, len(allSecrets))
	}
}

func TestInMemoryStore_ConcurrentPolicyOperations(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	policiesPerGoroutine := 3

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < policiesPerGoroutine; j++ {
				policyID := fmt.Sprintf("concurrent-policy-%d-%d", goroutineID, j)
				policy := data.Policy{
					ID:              policyID,
					Name:            fmt.Sprintf("Concurrent Policy %d-%d", goroutineID, j),
					SPIFFEIDPattern: fmt.Sprintf("spiffe://example\\.org/goroutine-%d/.*$", goroutineID),
					PathPattern:     fmt.Sprintf("concurrent/%d/*", goroutineID),
					Permissions:     []data.PolicyPermission{data.PermissionRead},
				}

				err := store.StorePolicy(ctx, policy)
				if err != nil {
					t.Errorf("Concurrent StorePolicy failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all policies were stored
	allPolicies, loadErr := store.LoadAllPolicies(ctx)
	if loadErr != nil {
		t.Errorf("LoadAllPolicies after concurrent writes failed: %v", loadErr)
	}

	expectedCount := numGoroutines * policiesPerGoroutine
	if len(allPolicies) != expectedCount {
		t.Errorf("Expected %d policies after concurrent writes, got %d",
			expectedCount, len(allPolicies))
	}
}

func TestInMemoryStore_MixedConcurrentOperations(t *testing.T) {
	testCipher := createTestCipher(t)
	store := NewInMemoryStore(testCipher, 10)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Concurrent secret operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			path := fmt.Sprintf("mixed/secret-%d", i)
			secret := kv.Value{
				Versions: map[int]kv.Version{
					1: {
						Data:    map[string]string{"key": fmt.Sprintf("value-%d", i)},
						Version: 1,
					},
				},
			}
			_ = store.StoreSecret(ctx, path, secret)
		}
	}()

	// Concurrent policy operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			policy := data.Policy{
				ID:              fmt.Sprintf("mixed-policy-%d", i),
				Name:            fmt.Sprintf("Mixed Policy %d", i),
				SPIFFEIDPattern: "^spiffe://example\\.org/mixed/.*$",
				PathPattern:     fmt.Sprintf("^mixed/%d/.*$", i),
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}
			_ = store.StorePolicy(ctx, policy)
		}
	}()

	// Concurrent read operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_, _ = store.LoadAllSecrets(ctx)
			_, _ = store.LoadAllPolicies(ctx)
		}
	}()

	wg.Wait()

	// Verify the final state
	secrets, loadSecretsErr := store.LoadAllSecrets(ctx)
	if loadSecretsErr != nil {
		t.Errorf("Final LoadAllSecrets failed: %v", loadSecretsErr)
	}

	policies, loadPoliciesErr := store.LoadAllPolicies(ctx)
	if loadPoliciesErr != nil {
		t.Errorf("Final LoadAllPolicies failed: %v", loadPoliciesErr)
	}

	if len(secrets) != 5 {
		t.Errorf("Expected 5 secrets, got %d", len(secrets))
	}

	if len(policies) != 5 {
		t.Errorf("Expected 5 policies, got %d", len(policies))
	}
}

func TestInMemoryStore_MaxVersionsConfig(t *testing.T) {
	testCipher := createTestCipher(t)
	maxVersions := 3
	store := NewInMemoryStore(testCipher, maxVersions)
	ctx := context.Background()

	// The kv.Config with MaxSecretVersions should be respected by the underlying kv.KV
	// This is more of an integration test to ensure the config is passed correctly

	// Store a secret (this tests that the KV was initialized with the config)
	secret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"test": "value1"},
				Version: 1,
			},
		},
	}

	storeErr := store.StoreSecret(ctx, "test/versions", secret)
	if storeErr != nil {
		t.Errorf("StoreSecret failed: %v", storeErr)
	}

	// Load it back to verify it worked
	loadedSecret, loadErr := store.LoadSecret(ctx, "test/versions")
	if loadErr != nil {
		t.Errorf("LoadSecret failed: %v", loadErr)
	}

	if loadedSecret == nil {
		t.Error("Expected non-nil loaded secret")
	}
}
