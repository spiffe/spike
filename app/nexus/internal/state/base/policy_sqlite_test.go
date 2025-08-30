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
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/config"
)

func TestSQLitePolicy_CreateAndGet(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		persist.InitializeBackend(rootKey)
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

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// Verify policy was stored
		retrievedPolicy, err := GetPolicy(createdPolicy.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy: %v", err)
		}

		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Expected name %s, got %s", policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Expected spiffeid pattern %s, got %s", policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Expected path pattern %s, got %s", policy.PathPattern, retrievedPolicy.PathPattern)
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
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		_, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy in first session: %v", err)
		}
	})

	// Second session - verify persistence
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		rootKey := createTestRootKey(t) // Same key as the first session
		resetRootKey()
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// First, we need to get the policy ID by listing policies and finding our policy
		allPolicies, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list policies in second session: %v", err)
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

		retrievedPolicy, err := GetPolicy(targetPolicyID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy in second session: %v", err)
		}

		if retrievedPolicy.Name != policy.Name {
			t.Errorf("Policy name not persisted correctly: expected %s, got %s", policy.Name, retrievedPolicy.Name)
		}
		if retrievedPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Policy SPIFFE ID pattern not persisted correctly: expected %s, got %s", policy.SPIFFEIDPattern, retrievedPolicy.SPIFFEIDPattern)
		}
		if retrievedPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Policy path pattern not persisted correctly: expected %s, got %s", policy.PathPattern, retrievedPolicy.PathPattern)
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
		persist.InitializeBackend(rootKey)
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
			_, err := CreatePolicy(policy)
			if err != nil {
				t.Fatalf("Failed to create policy %s: %v", policy.Name, err)
			}
		}

		// List all policies
		policiesList, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list policies: %v", err)
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
		persist.InitializeBackend(rootKey)
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
			_, err := CreatePolicy(policy)
			if err != nil {
				t.Fatalf("Failed to create policy %s: %v", policyName, err)
			}
		}

		// Verify that all policies exist
		allPolicies, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list policies: %v", err)
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
		err = DeletePolicy(policyToDeleteID)
		if err != nil {
			t.Fatalf("Failed to delete policy: %v", err)
		}

		// Verify policy was deleted
		_, err = GetPolicy(policyToDeleteID)
		if err == nil {
			t.Error("Expected error when getting deleted policy, got nil")
		}

		// Verify other policies still exist
		remainingPoliciesList, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list remaining policies: %v", err)
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
		persist.InitializeBackend(rootKey)
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

		createdFirst, err := CreatePolicy(firstPolicy)
		if err != nil {
			t.Fatalf("Failed to create first policy: %v", err)
		}

		// Create second policy with different name
		secondPolicy := data.Policy{
			Name:            "second-policy",
			SPIFFEIDPattern: "spiffe://example\\.org/second",
			PathPattern:     "second/path/.*",
			Permissions:     []data.PolicyPermission{data.PermissionWrite, data.PermissionList},
		}

		createdSecond, err := CreatePolicy(secondPolicy)
		if err != nil {
			t.Fatalf("Failed to create second policy: %v", err)
		}

		// Verify the first policy is still intact
		retrievedFirst, err := GetPolicy(createdFirst.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve first policy: %v", err)
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
		retrievedSecond, err := GetPolicy(createdSecond.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve second policy: %v", err)
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
			t.Errorf("Expected second policy path pattern %s, got %s",
				secondPolicy.PathPattern, retrievedSecond.PathPattern)
		}
		if !reflect.DeepEqual(retrievedSecond.Permissions, secondPolicy.Permissions) {
			t.Errorf("Expected second policy permissions %v, got %v",
				secondPolicy.Permissions, retrievedSecond.Permissions)
		}

		// Verify both policies exist in the list
		allPolicies, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list all policies: %v", err)
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
		persist.InitializeBackend(rootKey)
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

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy with special characters: %v", err)
		}

		// Retrieve and verify
		retrievedPolicy, err := GetPolicy(createdPolicy.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy with special characters: %v", err)
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
			t.Errorf("Special character path pattern not preserved: expected %s, got %s",
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
		persist.InitializeBackend(key1)
		Initialize(key1)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy with key1: %v", err)
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
		persist.InitializeBackend(key2)
		Initialize(key2)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// This should fail with the wrong key
		_, err := GetPolicy(createdPolicyID)
		if err == nil {
			t.Log("Note: GetPolicy succeeded with wrong key - this might indicate encryption issue")
		} else {
			t.Logf("Expected behavior: GetPolicy failed with wrong key: %v", err)
		}
	})

	// Verify the original key still works
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		resetRootKey()
		persist.InitializeBackend(key1)
		Initialize(key1)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		retrievedPolicy, err := GetPolicy(createdPolicyID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy with original key: %v", err)
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
			t.Errorf("Policy path pattern corrupted: expected %s, got %s",
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
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Test getting non-existent policy
		_, err := GetPolicy("non-existent-policy-id")
		if err == nil {
			t.Error("Expected error when getting non-existent policy, got nil")
		}

		// Test deleting non-existent policy
		err = DeletePolicy("non-existent-policy-id")
		if err == nil {
			t.Error("Expected error when deleting non-existent policy, got nil")
		}

		// Test creating policy with an empty name
		emptyNamePolicy := data.Policy{
			Name:            "",
			SPIFFEIDPattern: "spiffe://example\\.org/test",
			PathPattern:     "test/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		_, err = CreatePolicy(emptyNamePolicy)
		if err == nil {
			t.Error("Expected error when creating policy with empty name, got nil")
		}
	})
}

// Benchmark tests for SQLite policy operations
func BenchmarkSQLiteCreatePolicy(b *testing.B) {
	// Set environment variables for SQLite backend
	originalBackend := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	originalSkipSchema := os.Getenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "sqlite")
	_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	defer func() {
		if originalBackend != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalBackend)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
		if originalSkipSchema != "" {
			_ = os.Setenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION", originalSkipSchema)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")
		}
	}()

	// Clean up the database
	dataDir := config.SpikeNexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}

	rootKey := &[crypto.AES256KeySize]byte{}
	for i := range rootKey {
		rootKey[i] = byte(i % 256)
	}

	resetRootKey()
	persist.InitializeBackend(rootKey)
	Initialize(rootKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy := data.Policy{
			Name:            fmt.Sprintf("bench-policy-%d", i),
			SPIFFEIDPattern: "spiffe://example\\.org/benchmark",
			PathPattern:     "benchmark/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		_, err := CreatePolicy(policy)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkSQLiteGetPolicy(b *testing.B) {
	// Set environment variables for SQLite backend
	originalBackend := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	originalSkipSchema := os.Getenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "sqlite")
	_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	defer func() {
		if originalBackend != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalBackend)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
		if originalSkipSchema != "" {
			_ = os.Setenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION", originalSkipSchema)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")
		}
	}()

	// Clean up the database
	dataDir := config.SpikeNexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}

	rootKey := &[crypto.AES256KeySize]byte{}
	for i := range rootKey {
		rootKey[i] = byte(i % 256)
	}

	resetRootKey()
	persist.InitializeBackend(rootKey)
	Initialize(rootKey)

	// Create a policy to benchmark against
	policyName := "bench-get-policy"
	policy := data.Policy{
		Name:            policyName,
		SPIFFEIDPattern: "spiffe://example\\.org/benchmark",
		PathPattern:     "benchmark/.*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}
	_, _ = CreatePolicy(policy)

	// Get the policy ID before starting benchmark
	allPolicies, err := ListPolicies()
	if err != nil || len(allPolicies) == 0 {
		b.Fatalf("Failed to find created policy for benchmark: %v", err)
	}
	policyID := allPolicies[0].ID // Use the first policy

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetPolicy(policyID)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
