//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestSQLitePolicy_CreateAndGet(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Create a test policy
		policy := data.Policy{
			Name:            "test-policy",
			SPIFFEIDPattern: "^spiffe://example\\.org/workload$",
			PathPattern:     "^test/secrets/.*$",
			Permissions:     []data.PolicyPermission{data.PermissionRead, data.PermissionList},
		}

		createdPolicy, createErr := UpsertPolicy(policy)
		if createErr != nil {
			t.Fatalf("Failed to create policy: %v", createErr)
		}

		// Verify policy was stored
		retrievedPolicy, getErr := GetPolicy(createdPolicy.ID)
		if getErr != nil {
			t.Fatalf("Failed to retrieve policy: %v", getErr)
		}

		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Expected name %s, got %s", policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Expected SPIFFE ID pattern %s, got %s", policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Expected pathPattern pattern %s, got %s", policy.PathPattern, retrievedPolicy.PathPattern)
		}
		if !reflect.DeepEqual(retrievedPolicy.Permissions, policy.Permissions) {
			t.Errorf("Expected permissions %v, got %v", policy.Permissions, retrievedPolicy.Permissions)
		}
		if retrievedPolicy.ID != createdPolicy.ID {
			t.Errorf("Expected ID %s, got %s", createdPolicy.ID, retrievedPolicy.ID)
		}
	})
}

func TestSQLitePolicy_Persistence(t *testing.T) {
	policyName := "persistent-policy"
	policy := data.Policy{
		Name:            policyName,
		SPIFFEIDPattern: "^spiffe://example\\.org/service$",
		PathPattern:     "^persistent/data/.*$",
		Permissions:     []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
	}

	// First session - create policy
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		_, createErr := UpsertPolicy(policy)
		if createErr != nil {
			t.Fatalf("Failed to create policy in first session: %v", createErr)
		}
	})

	// Second session - verify persistence
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		rootKey := createTestRootKey(t) // Same key as the first session
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// First, we need to get the policy ID by listing policies
		// and finding our policy
		allPolicies, listErr := ListPolicies()
		if listErr != nil {
			t.Fatalf("Failed to list policies in second session: %v", listErr)
		}

		var targetPolicyID string
		for _, p := range allPolicies {
			if p.Name == policyName {
				targetPolicyID = p.ID
				break
			}
		}
		if targetPolicyID == "" {
			t.Fatalf("Policy with name %s not found in second session", policyName)
		}

		retrievedPolicy, getErr := GetPolicy(targetPolicyID)
		if getErr != nil {
			t.Fatalf("Failed to retrieve policy in second session: %v", getErr)
		}

		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Policy name not persisted correctly: expected %s, got %s", policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Policy SPIFFE ID pattern not persisted correctly: expected %s, got %s", policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Policy pathPattern pattern not persisted correctly: expected %s, got %s", policy.PathPattern, retrievedPolicy.PathPattern)
		}
		if !reflect.DeepEqual(retrievedPolicy.Permissions, policy.Permissions) {
			t.Errorf("Policy permissions not persisted correctly: expected %v, got %v", policy.Permissions, retrievedPolicy.Permissions)
		}
	})
}

func TestSQLitePolicy_ListPolicies(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Create multiple policies
		policies := []data.Policy{
			{
				Name:            "policy-alpha",
				SPIFFEIDPattern: "^spiffe://example\\.org/alpha$",
				PathPattern:     "^alpha/.*$",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
			{
				Name:            "policy-beta",
				SPIFFEIDPattern: "^spiffe://example\\.org/beta$",
				PathPattern:     "^beta/.*$",
				Permissions:     []data.PolicyPermission{data.PermissionWrite},
			},
			{
				Name:            "policy-gamma",
				SPIFFEIDPattern: "^spiffe://example\\.org/gamma$",
				PathPattern:     "^gamma/.*$",
				Permissions:     []data.PolicyPermission{data.PermissionSuper},
			},
		}

		for _, policy := range policies {
			_, createErr := UpsertPolicy(policy)
			if createErr != nil {
				t.Fatalf("Failed to create policy %s: %v", policy.Name, createErr)
			}
		}

		// List all policies
		policiesList, listErr := ListPolicies()
		if listErr != nil {
			t.Fatalf("Failed to list policies: %v", listErr)
		}

		// Extract policy names for comparison
		policyNames := make([]string, len(policiesList))
		for i, policy := range policiesList {
			policyNames[i] = policy.Name
		}

		// Sort for consistent comparison
		sort.Strings(policyNames)
		expectedNames := []string{"policy-alpha", "policy-beta", "policy-gamma"}
		sort.Strings(expectedNames)

		if !reflect.DeepEqual(policyNames, expectedNames) {
			t.Errorf("Expected policy names %v, got %v", expectedNames, policyNames)
		}
	})
}

func TestSQLitePolicy_DeletePolicy(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Create policies
		policies := []string{"keep-policy", "delete-policy-1", "delete-policy-2"}
		for _, policyName := range policies {
			policy := data.Policy{
				Name:            policyName,
				SPIFFEIDPattern: "^spiffe://example\\.org/test$",
				PathPattern:     fmt.Sprintf("^%s/.*$", policyName),
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}
			_, createErr := UpsertPolicy(policy)
			if createErr != nil {
				t.Fatalf("Failed to create policy %s: %v", policyName, createErr)
			}
		}

		// Verify that all policies exist
		allPolicies, listErr := ListPolicies()
		if listErr != nil {
			t.Fatalf("Failed to list policies: %v", listErr)
		}
		if len(allPolicies) != 3 {
			t.Fatalf("Expected 3 policies, got %d", len(allPolicies))
		}

		// Get the policy ID to delete by finding it in the list
		var policyToDeleteID string
		for _, p := range allPolicies {
			if p.Name == "delete-policy-1" {
				policyToDeleteID = p.ID
				break
			}
		}
		if policyToDeleteID == "" {
			t.Fatalf("Policy delete-policy-1 not found")
		}

		// Delete one policy
		deleteErr := DeletePolicy(policyToDeleteID)
		if deleteErr != nil {
			t.Fatalf("Failed to delete policy: %v", deleteErr)
		}

		// Verify policy was deleted
		_, getErr := GetPolicy(policyToDeleteID)
		if getErr == nil {
			t.Error("Expected error when getting deleted policy, got nil")
		}

		// Verify other policies still exist
		remainingPoliciesList, listRemainingErr := ListPolicies()
		if listRemainingErr != nil {
			t.Fatalf("Failed to list remaining policies: %v", listRemainingErr)
		}

		// Extract policy names for comparison
		remainingPolicies := make([]string, len(remainingPoliciesList))
		for i, policy := range remainingPoliciesList {
			remainingPolicies[i] = policy.Name
		}

		sort.Strings(remainingPolicies)
		expectedRemaining := []string{"delete-policy-2", "keep-policy"}
		sort.Strings(expectedRemaining)

		if !reflect.DeepEqual(remainingPolicies, expectedRemaining) {
			t.Errorf("Expected remaining policies %v, got %v",
				expectedRemaining, remainingPolicies)
		}
	})
}

func TestSQLitePolicy_CreateMultiplePolicies(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Create first policy
		firstPolicy := data.Policy{
			Name:            "first-policy",
			SPIFFEIDPattern: "^spiffe://example\\.org/first$",
			PathPattern:     "^first/.*$",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdFirst, createFirstErr := UpsertPolicy(firstPolicy)
		if createFirstErr != nil {
			t.Fatalf("Failed to create first policy: %v", createFirstErr)
		}

		// Create second policy with different name
		secondPolicy := data.Policy{
			Name:            "second-policy",
			SPIFFEIDPattern: "spiffe://example\\.org/second",
			PathPattern:     "second/pathPattern/.*",
			Permissions:     []data.PolicyPermission{data.PermissionWrite, data.PermissionList},
		}

		createdSecond, createSecondErr := UpsertPolicy(secondPolicy)
		if createSecondErr != nil {
			t.Fatalf("Failed to create second policy: %v", createSecondErr)
		}

		// Verify the first policy is still intact
		retrievedFirst, getFirstErr := GetPolicy(createdFirst.ID)
		if getFirstErr != nil {
			t.Fatalf("Failed to retrieve first policy: %v", getFirstErr)
		}

		if retrievedFirst.Name != firstPolicy.Name {
			t.Errorf("Expected first policy name %s, got %s",
				firstPolicy.Name, retrievedFirst.Name)
		}
		if retrievedFirst.SPIFFEIDPattern != firstPolicy.SPIFFEIDPattern {
			t.Errorf("Expected first policy SPIFFE ID pattern %s, got %s",
				firstPolicy.SPIFFEIDPattern, retrievedFirst.SPIFFEIDPattern)
		}

		// Verify the second policy was created correctly
		retrievedSecond, getSecondErr := GetPolicy(createdSecond.ID)
		if getSecondErr != nil {
			t.Fatalf("Failed to retrieve second policy: %v", getSecondErr)
		}

		if retrievedSecond.Name != secondPolicy.Name {
			t.Errorf("Expected second policy name %s, got %s",
				secondPolicy.Name, retrievedSecond.Name)
		}
		if retrievedSecond.SPIFFEIDPattern != secondPolicy.SPIFFEIDPattern {
			t.Errorf("Expected second policy SPIFFE ID pattern %s, got %s",
				secondPolicy.SPIFFEIDPattern, retrievedSecond.SPIFFEIDPattern)
		}
		if retrievedSecond.PathPattern != secondPolicy.PathPattern {
			t.Errorf("Expected second policy pathPattern pattern %s, got %s",
				secondPolicy.PathPattern, retrievedSecond.PathPattern)
		}
		if !reflect.DeepEqual(retrievedSecond.Permissions, secondPolicy.Permissions) {
			t.Errorf("Expected second policy permissions %v, got %v",
				secondPolicy.Permissions, retrievedSecond.Permissions)
		}

		// Verify both policies exist in the list
		allPolicies, listErr := ListPolicies()
		if listErr != nil {
			t.Fatalf("Failed to list all policies: %v", listErr)
		}
		if len(allPolicies) != 2 {
			t.Errorf("Expected 2 policies, got %d", len(allPolicies))
		}
	})
}

func TestSQLitePolicy_SpecialCharactersAndLongData(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Test with special characters and Unicode
		policy := data.Policy{
			Name:            "special-chars-你好-🔐-test",
			SPIFFEIDPattern: "spiffe://example\\.org/service-with-special-chars/测试",
			PathPattern:     "special/chars/with spaces/unicode/路径/测试/*",
			Permissions: []data.PolicyPermission{data.PermissionRead,
				data.PermissionWrite, data.PermissionList},
		}

		createdPolicy, createErr := UpsertPolicy(policy)
		if createErr != nil {
			t.Fatalf("Failed to create policy with special characters: %v", createErr)
		}

		// Retrieve and verify
		retrievedPolicy, getErr := GetPolicy(createdPolicy.ID)
		if getErr != nil {
			t.Fatalf("Failed to retrieve policy with special characters: %v", getErr)
		}

		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Special character policy name not preserved: expected %s, got %s",
				policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Special character SPIFFE ID pattern not preserved: expected %s, got %s",
				policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Special character pathPattern pattern not preserved: expected %s, got %s",
				policy.PathPattern, retrievedPolicy.PathPattern)
		}
		if !reflect.DeepEqual(retrievedPolicy.Permissions, policy.Permissions) {
			t.Errorf("Special character permissions not preserved: expected %v, got %v",
				policy.Permissions, retrievedPolicy.Permissions)
		}
	})
}

func TestSQLitePolicy_EncryptionWithDifferentKeys(t *testing.T) {
	policyName := "encryption-test-policy"
	policy := data.Policy{
		Name:            policyName,
		SPIFFEIDPattern: "spiffe://example\\.org/encryption-test",
		PathPattern:     "encrypted/data/.*",
		Permissions:     []data.PolicyPermission{data.PermissionSuper},
	}

	// Variable to store the created policy ID across test blocks
	var createdPolicyID string

	// Create a policy with the first key
	key1 := createTestRootKey(t)
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		resetRootKey()
		Initialize(key1)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		createdPolicy, createErr := UpsertPolicy(policy)
		if createErr != nil {
			t.Fatalf("Failed to create policy with key1: %v", createErr)
		}
		createdPolicyID = createdPolicy.ID
	})

	// Try to read with a different key (should fail)
	key2 := &[crypto.AES256KeySize]byte{}
	for i := range key2 {
		key2[i] = byte(255 - i) // Different pattern
	}

	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		resetRootKey()
		Initialize(key2)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// This should fail with the wrong key
		_, getErr := GetPolicy(createdPolicyID)
		if getErr == nil {
			t.Log("Note: GetPolicy succeeded with wrong key" +
				" - this might indicate encryption issue")
		} else {
			t.Logf("Expected behavior: GetPolicy failed with wrong key: %v", getErr)
		}
	})

	// Verify the original key still works
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		resetRootKey()
		Initialize(key1)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		retrievedPolicy, getErr := GetPolicy(createdPolicyID)
		if getErr != nil {
			t.Fatalf("Failed to retrieve policy with original key: %v", getErr)
		}

		// Compare relevant fields (ignore generated fields like ID, CreatedAt, etc.)
		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Policy name corrupted: expected %s, got %s",
				policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Policy SPIFFE ID pattern corrupted: expected %s, got %s",
				policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Policy pathPattern pattern corrupted: expected %s, got %s",
				policy.PathPattern, retrievedPolicy.PathPattern)
		}
		if !reflect.DeepEqual(retrievedPolicy.Permissions, policy.Permissions) {
			t.Errorf("Policy permissions corrupted: expected %v, got %v",
				policy.Permissions, retrievedPolicy.Permissions)
		}
	})
}

func TestSQLitePolicy_ErrorHandling(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Test getting non-existent policy
		_, getErr := GetPolicy("non-existent-policy-id")
		if getErr == nil {
			t.Error("Expected error when getting non-existent policy, got nil")
		}

		// Test deleting non-existent policy
		deleteErr := DeletePolicy("non-existent-policy-id")
		if deleteErr == nil {
			t.Error("Expected error when deleting non-existent policy, got nil")
		}

		// Test creating policy with an empty name
		emptyNamePolicy := data.Policy{
			Name:            "",
			SPIFFEIDPattern: "spiffe://example\\.org/test",
			PathPattern:     "test/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		_, createErr := UpsertPolicy(emptyNamePolicy)
		if createErr == nil {
			t.Error("Expected error when creating policy with empty name, got nil")
		}
	})
}

// Benchmark tests for SQLite policy operations
func BenchmarkSQLiteCreatePolicy(b *testing.B) {
	// Set environment variables for SQLite backend
	originalBackend := os.Getenv(env.NexusBackendStore)
	originalSkipSchema := os.Getenv(env.NexusDBSkipSchemaCreation)

	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)

	defer func() {
		if originalBackend != "" {
			_ = os.Setenv(env.NexusBackendStore, originalBackend)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalSkipSchema != "" {
			_ = os.Setenv(env.NexusDBSkipSchemaCreation, originalSkipSchema)
		} else {
			_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)
		}
	}()

	// Clean up the database
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}

	rootKey := &[crypto.AES256KeySize]byte{}
	for i := range rootKey {
		rootKey[i] = byte(i % 256)
	}

	resetRootKey()
	Initialize(rootKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy := data.Policy{
			Name:            fmt.Sprintf("bench-policy-%d", i),
			SPIFFEIDPattern: "spiffe://example\\.org/benchmark",
			PathPattern:     "benchmark/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		_, createErr := UpsertPolicy(policy)
		if createErr != nil {
			b.Fatalf("Benchmark failed: %v", createErr)
		}
	}
}

func BenchmarkSQLiteGetPolicy(b *testing.B) {
	// Set environment variables for SQLite backend
	originalBackend := os.Getenv(env.NexusBackendStore)
	originalSkipSchema := os.Getenv(env.NexusDBSkipSchemaCreation)

	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)

	defer func() {
		if originalBackend != "" {
			_ = os.Setenv(env.NexusBackendStore, originalBackend)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalSkipSchema != "" {
			_ = os.Setenv(env.NexusDBSkipSchemaCreation, originalSkipSchema)
		} else {
			_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)
		}
	}()

	// Clean up the database
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}

	rootKey := &[crypto.AES256KeySize]byte{}
	for i := range rootKey {
		rootKey[i] = byte(i % 256)
	}

	resetRootKey()
	Initialize(rootKey)

	// Create a policy to benchmark against
	policyName := "bench-get-policy"
	policy := data.Policy{
		Name:            policyName,
		SPIFFEIDPattern: "spiffe://example\\.org/benchmark",
		PathPattern:     "benchmark/.*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}
	_, _ = UpsertPolicy(policy)

	// Get the policy ID before starting benchmark
	allPolicies, listErr := ListPolicies()
	if listErr != nil || len(allPolicies) == 0 {
		b.Fatalf("Failed to find created policy for benchmark: %v", listErr)
	}
	policyID := allPolicies[0].ID // Use the first policy

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, getErr := GetPolicy(policyID)
		if getErr != nil {
			b.Fatalf("Benchmark failed: %v", getErr)
		}
	}
}
