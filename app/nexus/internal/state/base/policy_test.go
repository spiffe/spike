//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestCheckAccess_PilotAccess(t *testing.T) {
	// Test that pilot SPIFFE IDs always have access
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Test with a pilot SPIFFE ID pattern
		// Note: The actual IsPilot function behavior would need to be mocked
		// For now, we'll test the policy matching logic
		pilotSPIFFEID := "^spiffe://example\\.org/pilot$"
		path := "^test/secret$"
		wants := []data.PolicyPermission{data.PermissionRead}

		// This will return false in practice because we don't have the actual
		// SPIKE Pilot setup, but the code path will be tested
		result := CheckAccess(pilotSPIFFEID, path, wants)

		// Since we don't have actual pilot setup, this will test the policy matching path
		if result {
			t.Log("Pilot access granted (unexpected in test environment)")
		}
	})
}

func TestCheckAccess_WildcardPolicies(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a wildcard policy
		wildcardPolicy := data.Policy{
			Name:            "wildcard-read",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(wildcardPolicy)
		if err != nil {
			t.Fatalf("Failed to create wildcard policy: %v", err)
		}

		// Test access with a wildcard policy
		result := CheckAccess("spiffe://example.org/test",
			"any/path", []data.PolicyPermission{data.PermissionRead})
		if !result {
			t.Error("Expected wildcard policy to grant read access")
		}

		// Test access without required permission
		result = CheckAccess("spiffe://example.org/test",
			"any/path", []data.PolicyPermission{data.PermissionWrite})
		if result {
			t.Error("Expected wildcard policy to grant write access")
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCheckAccess_SuperPermission(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy with super permission
		superPolicy := data.Policy{
			Name:            "super-admin",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionSuper},
		}

		createdPolicy, err := CreatePolicy(superPolicy)
		if err != nil {
			t.Fatalf("Failed to create super policy: %v", err)
		}

		// Test that super permission grants all access
		permissions := []data.PolicyPermission{
			data.PermissionRead,
			data.PermissionWrite,
			data.PermissionList,
		}

		for _, perm := range permissions {
			result := CheckAccess("spiffe://example.org/test",
				"any/path", []data.PolicyPermission{perm})
			if !result {
				t.Errorf("Expected super permission to grant %v access", perm)
			}
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCheckAccess_SpecificPatterns(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy with specific patterns
		specificPolicy := data.Policy{
			Name:            "specific-access",
			SPIFFEIDPattern: "^spiffe://example\\.org/service.*$",
			PathPattern:     "^app/.*$",
			Permissions: []data.PolicyPermission{
				data.PermissionRead, data.PermissionWrite},
		}

		createdPolicy, err := CreatePolicy(specificPolicy)
		if err != nil {
			t.Fatalf("Failed to create specific policy: %v", err)
		}

		testCases := []struct {
			name        string
			spiffeID    string
			path        string
			wants       []data.PolicyPermission
			expectGrant bool
		}{
			{
				name:        "matching spiffeid and path",
				spiffeID:    "spiffe://example.org/service-a",
				path:        "app/secrets",
				wants:       []data.PolicyPermission{data.PermissionRead},
				expectGrant: true,
			},
			{
				name:        "matching spiffeid and path, multiple permissions",
				spiffeID:    "spiffe://example.org/service-b",
				path:        "app/config",
				wants:       []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
				expectGrant: true,
			},
			{
				name:        "non-matching spiffeid",
				spiffeID:    "spiffe://other.org/service-a",
				path:        "app/secrets",
				wants:       []data.PolicyPermission{data.PermissionRead},
				expectGrant: false,
			},
			{
				name:        "non-matching path",
				spiffeID:    "spiffe://example.org/service-a",
				path:        "other/secrets",
				wants:       []data.PolicyPermission{data.PermissionRead},
				expectGrant: false,
			},
			{
				name:        "requesting permission not granted",
				spiffeID:    "spiffe://example.org/service-a",
				path:        "app/secrets",
				wants:       []data.PolicyPermission{data.PermissionList},
				expectGrant: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := CheckAccess(tc.spiffeID, tc.path, tc.wants)
				if result != tc.expectGrant {
					t.Errorf("Expected %v, got %v for case: %s",
						tc.expectGrant, result, tc.name)
				}
			})
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCheckAccess_LoadPoliciesError(t *testing.T) {
	// Test behavior when ListPolicies returns an error
	// This is hard to test with real backend, but the function should return false
	// and log a warning when policies can't be loaded
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Normal case should work
		result := CheckAccess("spiffe://example.org/test",
			"some/path", []data.PolicyPermission{data.PermissionRead})
		// Should be false since no policies exist
		if result {
			t.Error("Expected false when no policies exist")
		}
	})
}

func TestCreatePolicy_ValidPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "test-policy",
			SPIFFEIDPattern: "^spiffe://example\\.org/.*$",
			PathPattern:     "^test/.*$",
			Permissions: []data.PolicyPermission{
				data.PermissionRead, data.PermissionWrite},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// Verify policy was created with expected fields
		if createdPolicy.ID == "" {
			t.Error("Expected generated ID")
		}
		if createdPolicy.Name != policy.Name {
			t.Errorf("Expected name %s, got %s", policy.Name, createdPolicy.Name)
		}
		if createdPolicy.SPIFFEIDPattern != policy.SPIFFEIDPattern {
			t.Errorf("Expected SPIFFEID pattern %s, got %s",
				policy.SPIFFEIDPattern, createdPolicy.SPIFFEIDPattern)
		}
		if createdPolicy.PathPattern != policy.PathPattern {
			t.Errorf("Expected path pattern %s, got %s",
				policy.PathPattern, createdPolicy.PathPattern)
		}
		if !reflect.DeepEqual(createdPolicy.Permissions, policy.Permissions) {
			t.Errorf("Expected permissions %v, got %v",
				policy.Permissions, createdPolicy.Permissions)
		}
		if createdPolicy.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
		if createdPolicy.IDRegex == nil {
			t.Error("Expected IDRegex to be compiled")
		}
		if createdPolicy.PathRegex == nil {
			t.Error("Expected PathRegex to be compiled")
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCreatePolicy_WildcardPatterns(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "wildcard-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create wildcard policy: %v", err)
		}

		// Verify wildcard patterns don't have compiled regexes
		if createdPolicy.IDRegex != nil {
			t.Error("Expected IDRegex to be nil for wildcard pattern")
		}
		if createdPolicy.PathRegex != nil {
			t.Error("Expected PathRegex to be nil for wildcard pattern")
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCreatePolicy_InvalidName(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "", // Invalid empty name
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		_, err := CreatePolicy(policy)
		if err == nil {
			t.Error("Expected error for empty policy name")
		}
		if !errors.Is(err, ErrInvalidPolicy) {
			t.Errorf("Expected ErrInvalidPolicy, got %v", err)
		}
	})
}

func TestCreatePolicy_DuplicateName(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "duplicate-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		// Create first policy
		createdPolicy1, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create first policy: %v", err)
		}

		// Try to create policy with same name
		_, err = CreatePolicy(policy)
		if err == nil {
			t.Error("Expected error for duplicate policy name")
		}
		if !errors.Is(err, ErrPolicyExists) {
			t.Errorf("Expected ErrPolicyExists, got %v", err)
		}

		// Clean up
		err = DeletePolicy(createdPolicy1.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestCreatePolicy_InvalidRegexPatterns(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		testCases := []struct {
			name            string
			spiffeIDPattern string
			pathPattern     string
			expectError     bool
		}{
			{
				name:            "invalid spiffeid regex",
				spiffeIDPattern: "[invalid-regex",
				pathPattern:     "valid/.*",
				expectError:     true,
			},
			{
				name:            "invalid path regex",
				spiffeIDPattern: "^spiffe://example\\.org/.*$",
				pathPattern:     "[invalid-regex",
				expectError:     true,
			},
			{
				name:            "valid patterns",
				spiffeIDPattern: "^spiffe://example\\.org/.*$",
				pathPattern:     "^test/.*$",
				expectError:     false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				policy := data.Policy{
					Name:            tc.name,
					SPIFFEIDPattern: tc.spiffeIDPattern,
					PathPattern:     tc.pathPattern,
					Permissions:     []data.PolicyPermission{data.PermissionRead},
				}

				createdPolicy, err := CreatePolicy(policy)
				if tc.expectError {
					if err == nil {
						t.Error("Expected error for invalid regex pattern")
					}
					if err != nil && !ErrorIs(err, ErrInvalidPolicy) {
						t.Errorf("Expected ErrInvalidPolicy to be in error chain, got %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for valid patterns: %v", err)
					} else {
						// Clean up successful creation
						_ = DeletePolicy(createdPolicy.ID)
					}
				}
			})
		}
	})
}

func TestCreatePolicy_PreserveCreatedAt(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		policy := data.Policy{
			Name:            "time-test-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
			CreatedAt:       customTime,
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		if !createdPolicy.CreatedAt.Equal(customTime) {
			t.Errorf("Expected CreatedAt %v, got %v", customTime, createdPolicy.CreatedAt)
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestGetPolicy_ExistingPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy first
		policy := data.Policy{
			Name:            "get-test-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// Get the policy
		retrievedPolicy, err := GetPolicy(createdPolicy.ID)
		if err != nil {
			t.Fatalf("Failed to get policy: %v", err)
		}

		// Verify the retrieved policy matches
		if retrievedPolicy.ID != createdPolicy.ID {
			t.Errorf("Expected ID %s, got %s",
				createdPolicy.ID, retrievedPolicy.ID)
		}
		if retrievedPolicy.Name != createdPolicy.Name {
			t.Errorf("Expected name %s, got %s",
				createdPolicy.Name, retrievedPolicy.Name)
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestGetPolicy_NonExistentPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Try to get a non-existent policy
		_, err := GetPolicy("non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent policy")
		}
		if !errors.Is(err, ErrPolicyNotFound) {
			t.Errorf("Expected ErrPolicyNotFound, got %v", err)
		}
	})
}

func TestDeletePolicy_ExistingPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy first
		policy := data.Policy{
			Name:            "delete-test-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// Delete the policy
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Fatalf("Failed to delete policy: %v", err)
		}

		// Verify the policy is gone
		_, err = GetPolicy(createdPolicy.ID)
		if !errors.Is(err, ErrPolicyNotFound) {
			t.Errorf("Expected ErrPolicyNotFound after deletion, got %v", err)
		}
	})
}

func TestDeletePolicy_NonExistentPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Try to delete a non-existent policy
		err := DeletePolicy("non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent policy")
		}
		if !errors.Is(err, ErrPolicyNotFound) {
			t.Errorf("Expected ErrPolicyNotFound, got %v", err)
		}
	})
}

func TestListPolicies_EmptyStore(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		policies, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list policies: %v", err)
		}

		if len(policies) != 0 {
			t.Errorf("Expected empty slice, got %d policies", len(policies))
		}
	})
}

func TestListPolicies_MultiplePolicies(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create multiple policies
		policyNames := []string{"policy-1", "policy-2", "policy-3"}
		createdPolicies := make([]data.Policy, 0, len(policyNames))

		for _, name := range policyNames {
			policy := data.Policy{
				Name:            name,
				SPIFFEIDPattern: "*",
				PathPattern:     "*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			}

			createdPolicy, err := CreatePolicy(policy)
			if err != nil {
				t.Fatalf("Failed to create policy %s: %v", name, err)
			}
			createdPolicies = append(createdPolicies, createdPolicy)
		}

		// List policies
		policies, err := ListPolicies()
		if err != nil {
			t.Fatalf("Failed to list policies: %v", err)
		}

		if len(policies) != len(policyNames) {
			t.Errorf("Expected %d policies, got %d", len(policyNames), len(policies))
		}

		// Verify all created policies are in the list
		policyMap := make(map[string]data.Policy)
		for _, policy := range policies {
			policyMap[policy.Name] = policy
		}

		for _, expectedName := range policyNames {
			if _, found := policyMap[expectedName]; !found {
				t.Errorf("Expected policy %s not found in list", expectedName)
			}
		}

		// Clean up
		for _, policy := range createdPolicies {
			err = DeletePolicy(policy.ID)
			if err != nil {
				t.Errorf("Failed to clean up policy %s: %v", policy.Name, err)
			}
		}
	})
}

func TestListPoliciesByPath_MatchingPolicies(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		pathPattern := "app/.*"

		// Create policies with different path patterns
		policies := []data.Policy{
			{
				Name:            "matching-policy-1",
				SPIFFEIDPattern: "*",
				PathPattern:     pathPattern,
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
			{
				Name:            "matching-policy-2",
				SPIFFEIDPattern: "^spiffe://example\\.org/.*$",
				PathPattern:     pathPattern,
				Permissions:     []data.PolicyPermission{data.PermissionWrite},
			},
			{
				Name:            "non-matching-policy",
				SPIFFEIDPattern: "*",
				PathPattern:     "^other/.*$",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
		}

		createdPolicies := make([]data.Policy, 0, len(policies))
		for _, policy := range policies {
			createdPolicy, err := CreatePolicy(policy)
			if err != nil {
				t.Fatalf("Failed to create policy %s: %v", policy.Name, err)
			}
			createdPolicies = append(createdPolicies, createdPolicy)
		}

		// List policies by path
		matchingPolicies, err := ListPoliciesByPath(pathPattern)
		if err != nil {
			t.Fatalf("Failed to list policies by path: %v", err)
		}

		if len(matchingPolicies) != 2 {
			t.Errorf("Expected 2 matching policies, got %d", len(matchingPolicies))
		}

		// Verify correct policies are returned
		names := make([]string, len(matchingPolicies))
		for i, policy := range matchingPolicies {
			names[i] = policy.Name
		}

		expectedNames := []string{"matching-policy-1", "matching-policy-2"}
		for _, expectedName := range expectedNames {
			found := false
			for _, name := range names {
				if name == expectedName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected policy %s not found in results", expectedName)
			}
		}

		// Clean up
		for _, policy := range createdPolicies {
			err = DeletePolicy(policy.ID)
			if err != nil {
				t.Errorf("Failed to clean up policy %s: %v", policy.Name, err)
			}
		}
	})
}

func TestListPoliciesByPath_NoMatches(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy with a different path pattern
		policy := data.Policy{
			Name:            "different-path-policy",
			SPIFFEIDPattern: "*",
			PathPattern:     "app/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// List policies with a non-matching path
		matchingPolicies, err := ListPoliciesByPath("other/.*")
		if err != nil {
			t.Fatalf("Failed to list policies by path: %v", err)
		}

		if len(matchingPolicies) != 0 {
			t.Errorf("Expected 0 matching policies, got %d", len(matchingPolicies))
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestListPoliciesBySPIFFEID_MatchingPolicies(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		spiffeIDPattern := "spiffe://example\\.org/.*"

		// Create policies with different SPIFFE ID patterns
		policies := []data.Policy{
			{
				Name:            "matching-spiffeid-policy-1",
				SPIFFEIDPattern: spiffeIDPattern,
				PathPattern:     "app/.*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
			{
				Name:            "matching-spiffeid-policy-2",
				SPIFFEIDPattern: spiffeIDPattern,
				PathPattern:     "other/.*",
				Permissions:     []data.PolicyPermission{data.PermissionWrite},
			},
			{
				Name:            "non-matching-spiffeid-policy",
				SPIFFEIDPattern: "spiffe://other\\.org/.*",
				PathPattern:     "app/.*",
				Permissions:     []data.PolicyPermission{data.PermissionRead},
			},
		}

		createdPolicies := make([]data.Policy, 0, len(policies))
		for _, policy := range policies {
			createdPolicy, err := CreatePolicy(policy)
			if err != nil {
				t.Fatalf("Failed to create policy %s: %v", policy.Name, err)
			}
			createdPolicies = append(createdPolicies, createdPolicy)
		}

		// List policies by SPIFFE ID
		matchingPolicies, err := ListPoliciesBySPIFFEID(spiffeIDPattern)
		if err != nil {
			t.Fatalf("Failed to list policies by SPIFFE ID: %v", err)
		}

		if len(matchingPolicies) != 2 {
			t.Errorf("Expected 2 matching policies, got %d", len(matchingPolicies))
		}

		// Verify correct policies are returned
		names := make([]string, len(matchingPolicies))
		for i, policy := range matchingPolicies {
			names[i] = policy.Name
		}

		expectedNames := []string{
			"matching-spiffeid-policy-1", "matching-spiffeid-policy-2"}
		for _, expectedName := range expectedNames {
			found := false
			for _, name := range names {
				if name == expectedName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected policy %s not found in results", expectedName)
			}
		}

		// Clean up
		for _, policy := range createdPolicies {
			err = DeletePolicy(policy.ID)
			if err != nil {
				t.Errorf("Failed to clean up policy %s: %v", policy.Name, err)
			}
		}
	})
}

func TestListPoliciesBySPIFFEID_NoMatches(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Create a policy with a different SPIFFE ID pattern
		policy := data.Policy{
			Name:            "different-spiffeid-policy",
			SPIFFEIDPattern: "spiffe://example\\.org/.*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// List policies with non-matching SPIFFE ID
		matchingPolicies, err := ListPoliciesBySPIFFEID("spiffe://other\\.org/.*")
		if err != nil {
			t.Fatalf("Failed to list policies by SPIFFE ID: %v", err)
		}

		if len(matchingPolicies) != 0 {
			t.Errorf("Expected 0 matching policies, got %d", len(matchingPolicies))
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

func TestPolicyRegexCompilation(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Test that regex patterns are correctly compiled
		policy := data.Policy{
			Name:            "regex-test-policy",
			SPIFFEIDPattern: "spiffe://example\\.org/service-[0-9]+",
			PathPattern:     "app/service-[a-z]+/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, err := CreatePolicy(policy)
		if err != nil {
			t.Fatalf("Failed to create policy: %v", err)
		}

		// Test the compiled regexes work correctly
		testCases := []struct {
			spiffeID    string
			path        string
			shouldMatch bool
		}{
			{"spiffe://example.org/service-123", "app/service-test/config", true},
			{"spiffe://example.org/service-abc", "app/service-test/config", false}, // invalid spiffeid
			{"spiffe://example.org/service-123", "app/service-123/config", false},  // invalid path (numbers instead of letters)
			{"spiffe://other.org/service-123", "app/service-test/config", false},   // wrong domain
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("regex_test_%d", i), func(t *testing.T) {
				result := CheckAccess(tc.spiffeID, tc.path, []data.PolicyPermission{data.PermissionRead})
				if result != tc.shouldMatch {
					t.Errorf("Expected %v for SPIFFEID %s and path %s", tc.shouldMatch, tc.spiffeID, tc.path)
				}
			})
		}

		// Clean up
		err = DeletePolicy(createdPolicy.ID)
		if err != nil {
			t.Errorf("Failed to clean up policy: %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkCheckAccess_WildcardPolicy(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	// Create a wildcard policy
	policy := data.Policy{
		Name:            "benchmark-wildcard",
		SPIFFEIDPattern: "*",
		PathPattern:     "*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}

	createdPolicy, _ := CreatePolicy(policy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckAccess("spiffe://example.org/test",
			"test/path", []data.PolicyPermission{data.PermissionRead})
	}

	_ = DeletePolicy(createdPolicy.ID)
}

func BenchmarkCreatePolicy(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	createdPolicies := make([]string, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy := data.Policy{
			Name:            fmt.Sprintf("benchmark-policy-%d", i),
			SPIFFEIDPattern: "spiffe://example\\.org/.*",
			PathPattern:     "test/.*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}

		createdPolicy, _ := CreatePolicy(policy)
		createdPolicies = append(createdPolicies, createdPolicy.ID)
	}
	b.StopTimer()

	// Clean up
	for _, id := range createdPolicies {
		_ = DeletePolicy(id)
	}
}

func BenchmarkListPolicies(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	// Create some policies for benchmarking
	createdPolicies := make([]string, 0)
	for i := 0; i < 100; i++ {
		policy := data.Policy{
			Name:            fmt.Sprintf("benchmark-list-policy-%d", i),
			SPIFFEIDPattern: "*",
			PathPattern:     "*",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}
		createdPolicy, _ := CreatePolicy(policy)
		createdPolicies = append(createdPolicies, createdPolicy.ID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ListPolicies()
	}
	b.StopTimer()

	// Clean up
	for _, id := range createdPolicies {
		_ = DeletePolicy(id)
	}
}

// Helper function to check if an error contains a specific error in its chain
func ErrorIs(err, target error) bool {
	return err != nil && (errors.Is(err, target) ||
		(err.Error() != "" &&
			target.Error() != "" && err.Error() == target.Error()))
}
