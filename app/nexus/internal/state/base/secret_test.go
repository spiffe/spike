//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestUpsertSecret_NewSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/new-secret"
		values := map[string]string{
			"username": "admin",
			"password": "secret123",
		}

		err := UpsertSecret(path, values)
		if err != nil {
			t.Fatalf("Failed to upsert new secret: %v", err)
		}

		// Verify the secret was created
		retrievedValues, err := GetSecret(path, 0) // 0 = current version
		if err != nil {
			t.Fatalf("Failed to retrieve secret: %v", err)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Expected values %v, got %v", values, retrievedValues)
		}

		// Verify metadata
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", err)
		}

		if rawSecret.Metadata.CurrentVersion != 1 {
			t.Errorf("Expected current version 1, got %d", rawSecret.Metadata.CurrentVersion)
		}
		if rawSecret.Metadata.OldestVersion != 1 {
			t.Errorf("Expected oldest version 1, got %d", rawSecret.Metadata.OldestVersion)
		}
		if len(rawSecret.Versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(rawSecret.Versions))
		}
	})
}

func TestUpsertSecret_ExistingSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/existing-secret"

		// Create initial version
		initialValues := map[string]string{
			"key1": "value1",
		}
		err := UpsertSecret(path, initialValues)
		if err != nil {
			t.Fatalf("Failed to create initial secret: %v", err)
		}

		// Update with new values
		updatedValues := map[string]string{
			"key1": "updated_value1",
			"key2": "value2",
		}
		err = UpsertSecret(path, updatedValues)
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}

		// Verify current version has updated values
		currentValues, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve current version: %v", err)
		}

		if !reflect.DeepEqual(currentValues, updatedValues) {
			t.Errorf("Expected current values %v, got %v", updatedValues, currentValues)
		}

		// Verify previous version still exists
		previousValues, err := GetSecret(path, 1)
		if err != nil {
			t.Fatalf("Failed to retrieve previous version: %v", err)
		}

		if !reflect.DeepEqual(previousValues, initialValues) {
			t.Errorf("Expected previous values %v, got %v", initialValues, previousValues)
		}

		// Verify metadata
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", err)
		}

		if rawSecret.Metadata.CurrentVersion != 2 {
			t.Errorf("Expected current version 2, got %d", rawSecret.Metadata.CurrentVersion)
		}
		if rawSecret.Metadata.OldestVersion != 1 {
			t.Errorf("Expected oldest version 1, got %d", rawSecret.Metadata.OldestVersion)
		}
	})
}

func TestUpsertSecret_VersionPruning(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		withEnvironment(t, "SPIKE_NEXUS_MAX_SECRET_VERSIONS", "3", func() {
			resetBackendForTest()
			persist.InitializeBackend(nil)

			path := "/test/version-pruning"

			// Create 5 versions (more than max of 3)
			for i := 1; i <= 5; i++ {
				values := map[string]string{
					"version": fmt.Sprintf("v%d", i),
				}
				err := UpsertSecret(path, values)
				if err != nil {
					t.Fatalf("Failed to create version %d: %v", i, err)
				}
			}

			// Verify only 3 versions remain (versions 3, 4, 5)
			rawSecret, err := GetRawSecret(path, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve raw secret: %v", err)
			}

			if len(rawSecret.Versions) != 3 {
				t.Errorf("Expected 3 versions after pruning, got %d", len(rawSecret.Versions))
			}

			// Verify oldest version is correct (should be version 3)
			if rawSecret.Metadata.OldestVersion != 3 {
				t.Errorf("Expected oldest version 3, got %d", rawSecret.Metadata.OldestVersion)
			}

			// Verify current version is correct (should be version 5)
			if rawSecret.Metadata.CurrentVersion != 5 {
				t.Errorf("Expected current version 5, got %d", rawSecret.Metadata.CurrentVersion)
			}

			// Verify old versions are gone
			for version := 1; version <= 2; version++ {
				_, err := GetSecret(path, version)
				if err == nil {
					t.Errorf("Expected version %d to be pruned", version)
				}
			}

			// Verify remaining versions exist
			for version := 3; version <= 5; version++ {
				values, err := GetSecret(path, version)
				if err != nil {
					t.Errorf("Version %d should exist: %v", version, err)
				}
				expectedValue := fmt.Sprintf("v%d", version)
				if values["version"] != expectedValue {
					t.Errorf("Expected version %d to have value %s, got %s", version, expectedValue, values["version"])
				}
			}
		})
	})
}

func TestDeleteSecret_CurrentVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/delete-current"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Delete current version (should be version 3)
		err := DeleteSecret(path, []int{0}) // 0 = current version
		if err != nil {
			t.Fatalf("Failed to delete current version: %v", err)
		}

		// Verify current version is now 2
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", err)
		}

		if rawSecret.Metadata.CurrentVersion != 2 {
			t.Errorf("Expected current version to be 2, got %d", rawSecret.Metadata.CurrentVersion)
		}

		// Verify version 3 is marked as deleted
		version3, exists := rawSecret.Versions[3]
		if !exists {
			t.Error("Version 3 should still exist but be marked as deleted")
		}
		if version3.DeletedTime == nil {
			t.Error("Version 3 should be marked as deleted")
		}

		// Verify version 2 is accessible
		values, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get current version: %v", err)
		}
		if values["version"] != "v2" {
			t.Errorf("Expected current version to be v2, got %s", values["version"])
		}
	})
}

func TestDeleteSecret_SpecificVersions(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/delete-specific"

		// Create multiple versions
		for i := 1; i <= 4; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Delete specific versions (1 and 3)
		err := DeleteSecret(path, []int{1, 3})
		if err != nil {
			t.Fatalf("Failed to delete specific versions: %v", err)
		}

		// Verify current version is still 4
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", err)
		}

		if rawSecret.Metadata.CurrentVersion != 4 {
			t.Errorf("Expected current version to remain 4, got %d", rawSecret.Metadata.CurrentVersion)
		}

		// Verify versions 1 and 3 are marked as deleted
		for _, version := range []int{1, 3} {
			v, exists := rawSecret.Versions[version]
			if !exists {
				t.Errorf("Version %d should still exist", version)
			}
			if v.DeletedTime == nil {
				t.Errorf("Version %d should be marked as deleted", version)
			}
		}

		// Verify versions 2 and 4 are still accessible
		for _, version := range []int{2, 4} {
			values, err := GetSecret(path, version)
			if err != nil {
				t.Errorf("Version %d should still be accessible: %v", version, err)
			}
			expectedValue := fmt.Sprintf("v%d", version)
			if values["version"] != expectedValue {
				t.Errorf("Expected version %d to have value %s", version, expectedValue)
			}
		}
	})
}

func TestDeleteSecret_AllVersions(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/delete-all"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Delete all versions
		err := DeleteSecret(path, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("Failed to delete all versions: %v", err)
		}

		// Verify current version is 0 (no active versions)
		_, err = GetRawSecret(path, 0)
		if err == nil {
			t.Error("Expected error when trying to get current version of fully deleted secret")
		}

		// Try to get the raw secret without version validation
		ctx := persist.Backend()
		secret, err := ctx.LoadSecret(nil, path)
		if err != nil {
			t.Fatalf("Failed to load raw secret: %v", err)
		}

		if secret.Metadata.CurrentVersion != 0 {
			t.Errorf("Expected current version to be 0, got %d", secret.Metadata.CurrentVersion)
		}

		// Verify all versions are marked as deleted
		for version := 1; version <= 3; version++ {
			v, exists := secret.Versions[version]
			if !exists {
				t.Errorf("Version %d should still exist", version)
			}
			if v.DeletedTime == nil {
				t.Errorf("Version %d should be marked as deleted", version)
			}
		}
	})
}

func TestDeleteSecret_NonExistentSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		err := DeleteSecret("/test/nonexistent", []int{1})
		if err == nil {
			t.Error("Expected error when deleting non-existent secret")
		}
	})
}

func TestUndeleteSecret_SpecificVersions(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/undelete-specific"

		// Create and then delete multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Delete versions 1 and 2
		err := DeleteSecret(path, []int{1, 2})
		if err != nil {
			t.Fatalf("Failed to delete versions: %v", err)
		}

		// Undelete version 1
		err = UndeleteSecret(path, []int{1})
		if err != nil {
			t.Fatalf("Failed to undelete version 1: %v", err)
		}

		// Verify version 1 is now accessible
		values, err := GetSecret(path, 1)
		if err != nil {
			t.Errorf("Version 1 should be accessible after undelete: %v", err)
		}
		if values["version"] != "v1" {
			t.Errorf("Expected version 1 to have value v1, got %s", values["version"])
		}

		// Verify version 2 is still deleted
		_, err = GetSecret(path, 2)
		if err == nil {
			t.Error("Version 2 should still be deleted")
		}

		// Verify current version is still 3
		currentValues, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get current version: %v", err)
		}
		if currentValues["version"] != "v3" {
			t.Errorf("Expected current version to be v3, got %s", currentValues["version"])
		}
	})
}

func TestUndeleteSecret_AllDeleted(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/undelete-all-deleted"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Delete all versions
		err := DeleteSecret(path, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("Failed to delete all versions: %v", err)
		}

		// Undelete without specifying versions (should undelete highest)
		err = UndeleteSecret(path, []int{})
		if err != nil {
			t.Fatalf("Failed to undelete: %v", err)
		}

		// Verify version 3 is current and accessible
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get raw secret: %v", err)
		}

		if rawSecret.Metadata.CurrentVersion != 3 {
			t.Errorf("Expected current version to be 3, got %d", rawSecret.Metadata.CurrentVersion)
		}

		values, err := GetSecret(path, 0)
		if err != nil {
			t.Errorf("Current version should be accessible: %v", err)
		}
		if values["version"] != "v3" {
			t.Errorf("Expected current version to be v3, got %s", values["version"])
		}
	})
}

func TestUndeleteSecret_NonExistentSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		err := UndeleteSecret("/test/nonexistent", []int{1})
		if err == nil {
			t.Error("Expected error when undeleting non-existent secret")
		}
	})
}

func TestGetSecret_CurrentVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/get-current"
		values := map[string]string{
			"key": "value",
		}

		err := UpsertSecret(path, values)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Get current version (version 0)
		retrievedValues, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get current version: %v", err)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Expected values %v, got %v", values, retrievedValues)
		}
	})
}

func TestGetSecret_SpecificVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/get-specific"

		// Create multiple versions
		expectedValues := make(map[int]map[string]string)
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
				"data":    fmt.Sprintf("data-%d", i),
			}
			expectedValues[i] = values

			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Retrieve each version specifically
		for version := 1; version <= 3; version++ {
			retrievedValues, err := GetSecret(path, version)
			if err != nil {
				t.Errorf("Failed to get version %d: %v", version, err)
				continue
			}

			if !reflect.DeepEqual(retrievedValues, expectedValues[version]) {
				t.Errorf("Version %d: expected %v, got %v", version, expectedValues[version], retrievedValues)
			}
		}
	})
}

func TestGetSecret_NonExistentSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		_, err := GetSecret("/test/nonexistent", 1)
		if err == nil {
			t.Error("Expected error when getting non-existent secret")
		}
	})
}

func TestGetSecret_NonExistentVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/nonexistent-version"
		values := map[string]string{
			"key": "value",
		}

		err := UpsertSecret(path, values)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Try to get non-existent version
		_, err = GetSecret(path, 999)
		if err == nil {
			t.Error("Expected error when getting non-existent version")
		}
	})
}

func TestGetSecret_DeletedVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/deleted-version"
		values := map[string]string{
			"key": "value",
		}

		err := UpsertSecret(path, values)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Delete version 1
		err = DeleteSecret(path, []int{1})
		if err != nil {
			t.Fatalf("Failed to delete version 1: %v", err)
		}

		// Try to get deleted version
		_, err = GetSecret(path, 1)
		if err == nil {
			t.Error("Expected error when getting deleted version")
		}
	})
}

func TestGetRawSecret_WithMetadata(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/raw-secret"

		// Create multiple versions
		for i := 1; i <= 2; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			err := UpsertSecret(path, values)
			if err != nil {
				t.Fatalf("Failed to create version %d: %v", i, err)
			}
		}

		// Get raw secret
		rawSecret, err := GetRawSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get raw secret: %v", err)
		}

		// Verify metadata
		if rawSecret.Metadata.CurrentVersion != 2 {
			t.Errorf("Expected current version 2, got %d", rawSecret.Metadata.CurrentVersion)
		}
		if rawSecret.Metadata.OldestVersion != 1 {
			t.Errorf("Expected oldest version 1, got %d", rawSecret.Metadata.OldestVersion)
		}
		if len(rawSecret.Versions) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(rawSecret.Versions))
		}

		// Verify version data
		v1, exists := rawSecret.Versions[1]
		if !exists {
			t.Error("Version 1 should exist")
		} else if v1.Data["version"] != "v1" {
			t.Errorf("Expected version 1 to have value v1, got %s", v1.Data["version"])
		}

		v2, exists := rawSecret.Versions[2]
		if !exists {
			t.Error("Version 2 should exist")
		} else if v2.Data["version"] != "v2" {
			t.Errorf("Expected version 2 to have value v2, got %s", v2.Data["version"])
		}
	})
}

func TestGetRawSecret_AllVersionsDeleted(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/all-deleted"
		values := map[string]string{
			"key": "value",
		}

		err := UpsertSecret(path, values)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Delete all versions
		err = DeleteSecret(path, []int{1})
		if err != nil {
			t.Fatalf("Failed to delete version: %v", err)
		}

		// Try to get raw secret when all versions are deleted
		_, err = GetRawSecret(path, 0)
		if err == nil {
			t.Error("Expected error when all versions are deleted")
		}
	})
}

func TestSecretOperations_ConcurrentAccess(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/concurrent"

		// Create initial secret
		err := UpsertSecret(path, map[string]string{"counter": "0"})
		if err != nil {
			t.Fatalf("Failed to create initial secret: %v", err)
		}

		// Test that multiple operations work correctly
		// Note: This is a simple test since the memory backend is not truly concurrent-safe
		// But it tests the API works correctly in sequence

		operations := []func() error{
			func() error {
				return UpsertSecret(path, map[string]string{"counter": "1"})
			},
			func() error {
				_, err := GetSecret(path, 0)
				return err
			},
			func() error {
				return DeleteSecret(path, []int{1})
			},
			func() error {
				return UndeleteSecret(path, []int{1})
			},
			func() error {
				_, err := GetRawSecret(path, 0)
				return err
			},
		}

		for i, op := range operations {
			if err := op(); err != nil {
				t.Errorf("Operation %d failed: %v", i, err)
			}
		}
	})
}

func TestSecretOperations_EmptyValues(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/empty-values"

		// Test with empty map
		err := UpsertSecret(path, map[string]string{})
		if err != nil {
			t.Fatalf("Failed to create secret with empty values: %v", err)
		}

		// Retrieve and verify
		values, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get secret with empty values: %v", err)
		}

		if len(values) != 0 {
			t.Errorf("Expected empty values map, got %v", values)
		}

		// Test with nil map
		err = UpsertSecret(path, nil)
		if err != nil {
			t.Fatalf("Failed to create secret with nil values: %v", err)
		}

		values, err = GetSecret(path, 0) // Should be version 2 now
		if err != nil {
			t.Fatalf("Failed to get secret with nil values: %v", err)
		}

		if values == nil {
			t.Error("Expected non-nil values map, got nil")
		}
		if len(values) != 0 {
			t.Errorf("Expected empty values map, got %v", values)
		}
	})
}

func TestSecretOperations_LargeValues(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/large-values"

		// Create large secret with many keys
		largeValues := make(map[string]string)
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%03d", i)
			value := fmt.Sprintf("value_%03d_with_some_longer_content_to_make_it_bigger", i)
			largeValues[key] = value
		}

		err := UpsertSecret(path, largeValues)
		if err != nil {
			t.Fatalf("Failed to create large secret: %v", err)
		}

		// Retrieve and verify
		retrievedValues, err := GetSecret(path, 0)
		if err != nil {
			t.Fatalf("Failed to get large secret: %v", err)
		}

		if len(retrievedValues) != len(largeValues) {
			t.Errorf("Expected %d keys, got %d", len(largeValues), len(retrievedValues))
		}

		if !reflect.DeepEqual(retrievedValues, largeValues) {
			t.Error("Retrieved large values don't match original")
		}
	})
}

func TestSecretOperations_SpecialCharacters(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Test paths with special characters
		specialPaths := []string{
			"/test/special-chars",
			"/test/with spaces",
			"/test/with.dots",
			"/test/with_underscores",
			"/test/with-dashes",
			"/test/with/deep/nested/path",
		}

		for _, path := range specialPaths {
			t.Run(path, func(t *testing.T) {
				values := map[string]string{
					"unicode":       "Test with üñíçödé characters 中文",
					"special_chars": "!@#$%^&*()_+-=[]{}|;:'\",.<>?",
					"newlines":      "Line 1\nLine 2\nLine 3",
					"tabs":          "Column1\tColumn2\tColumn3",
					"quotes":        `Both "double" and 'single' quotes`,
				}

				err := UpsertSecret(path, values)
				if err != nil {
					t.Fatalf("Failed to create secret with special characters: %v", err)
				}

				retrievedValues, err := GetSecret(path, 0)
				if err != nil {
					t.Fatalf("Failed to get secret with special characters: %v", err)
				}

				if !reflect.DeepEqual(retrievedValues, values) {
					t.Error("Retrieved special character values don't match original")
				}
			})
		}
	})
}

// Benchmark tests
func BenchmarkUpsertSecret_NewSecret(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/bench/secret-%d", i)
		UpsertSecret(path, values)
	}
}

func BenchmarkUpsertSecret_UpdateExisting(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/update-secret"
	initialValues := map[string]string{
		"username": "admin",
		"password": "initial",
	}

	// Create initial secret
	UpsertSecret(path, initialValues)

	updatedValues := map[string]string{
		"username": "admin",
		"password": "updated",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updatedValues["counter"] = fmt.Sprintf("%d", i)
		UpsertSecret(path, updatedValues)
	}
}

func BenchmarkGetSecret(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/get-secret"
	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}

	UpsertSecret(path, values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSecret(path, 0)
	}
}

func BenchmarkGetRawSecret(b *testing.B) {
	original := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")
	defer func() {
		if original != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", original)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/get-raw-secret"
	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}

	UpsertSecret(path, values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRawSecret(path, 0)
	}
}

// Helper function to manage environment variables in tests
func withEnvironment32(t *testing.T, key, value string, testFunc func()) {
	original := os.Getenv(key)
	os.Setenv(key, value)
	defer func() {
		if original != "" {
			os.Setenv(key, original)
		} else {
			os.Unsetenv(key)
		}
	}()
	testFunc()
}

// Helper function to reset backend state for testing
func resetBackendForTest44() {
	// This will be implemented to ensure clean state between tests
	// For now, we rely on initializing a fresh memory backend
}
