//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	appEnv "github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestUpsertSecret_NewSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/new-secret"
		values := map[string]string{
			"username": "admin",
			"password": "secret123",
		}

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to upsert new secret: %v", upsertErr)
		}

		// Verify the secret was created
		retrievedValues, getErr := GetSecret(path, 0) // 0 = current version
		if getErr != nil {
			t.Fatalf("Failed to retrieve secret: %v", getErr)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Expected values %v, got %v", values, retrievedValues)
		}

		// Verify metadata
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", getRawErr)
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

		path := "test/existing-secret"

		// Create the initial version
		initialValues := map[string]string{
			"key1": "value1",
		}
		createErr := UpsertSecret(path, initialValues)
		if createErr != nil {
			t.Fatalf("Failed to create initial secret: %v", createErr)
		}

		// Update with new values
		updatedValues := map[string]string{
			"key1": "updated_value1",
			"key2": "value2",
		}
		updateErr := UpsertSecret(path, updatedValues)
		if updateErr != nil {
			t.Fatalf("Failed to update secret: %v", updateErr)
		}

		// Verify the current version has updated values
		currentValues, getCurrentErr := GetSecret(path, 0)
		if getCurrentErr != nil {
			t.Fatalf("Failed to retrieve current version: %v", getCurrentErr)
		}

		if !reflect.DeepEqual(currentValues, updatedValues) {
			t.Errorf("Expected current values %v, got %v", updatedValues, currentValues)
		}

		// Verify the previous version still exists
		previousValues, getPrevErr := GetSecret(path, 1)
		if getPrevErr != nil {
			t.Fatalf("Failed to retrieve previous version: %v", getPrevErr)
		}

		if !reflect.DeepEqual(previousValues, initialValues) {
			t.Errorf("Expected previous values %v, got %v", initialValues, previousValues)
		}

		// Verify metadata
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", getRawErr)
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

			path := "test/version-pruning"

			// Create 5 versions (more than max of 3)
			for i := 1; i <= 5; i++ {
				versionValues := map[string]string{
					"version": fmt.Sprintf("v%d", i),
				}
				upsertErr := UpsertSecret(path, versionValues)
				if upsertErr != nil {
					t.Fatalf("Failed to create version %d: %v", i, upsertErr)
				}
			}

			// Verify only 3 versions remain (versions 3, 4, 5)
			rawSecret, getRawErr := GetRawSecret(path, 0)
			if getRawErr != nil {
				t.Fatalf("Failed to retrieve raw secret: %v", getRawErr)
			}

			if len(rawSecret.Versions) != 3 {
				t.Errorf("Expected 3 versions after pruning, got %d", len(rawSecret.Versions))
			}

			// Verify the oldest version is correct (should be version 3)
			if rawSecret.Metadata.OldestVersion != 3 {
				t.Errorf("Expected oldest version 3, got %d", rawSecret.Metadata.OldestVersion)
			}

			// Verify the current version is correct (should be version 5)
			if rawSecret.Metadata.CurrentVersion != 5 {
				t.Errorf("Expected current version 5, got %d", rawSecret.Metadata.CurrentVersion)
			}

			// Verify old versions are gone
			for version := 1; version <= 2; version++ {
				_, getErr := GetSecret(path, version)
				if getErr == nil {
					t.Errorf("Expected version %d to be pruned", version)
				}
			}

			// Verify remaining versions exist
			for version := 3; version <= 5; version++ {
				values, getErr := GetSecret(path, version)
				if getErr != nil {
					t.Errorf("Version %d should exist: %v", version, getErr)
				}
				expectedValue := fmt.Sprintf("v%d", version)
				if values["version"] != expectedValue {
					t.Errorf(
						"Expected version %d to have value %s, got %s",
						version, expectedValue, values["version"],
					)
				}
			}
		})
	})
}

func TestDeleteSecret_CurrentVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/delete-current"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			versionValues := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, versionValues)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Delete the current version (should be version 3)
		deleteErr := DeleteSecret(path, []int{0}) // 0 = current version
		if deleteErr != nil {
			t.Fatalf("Failed to delete current version: %v", deleteErr)
		}

		// Verify current version is now 2
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", getRawErr)
		}

		if rawSecret.Metadata.CurrentVersion != 2 {
			t.Errorf(
				"Expected current version to be 2, got %d",
				rawSecret.Metadata.CurrentVersion,
			)
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
		values, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to get current version: %v", getErr)
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

		path := "test/delete-specific"

		// Create multiple versions
		for i := 1; i <= 4; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Delete specific versions (1 and 3)
		deleteErr := DeleteSecret(path, []int{1, 3})
		if deleteErr != nil {
			t.Fatalf("Failed to delete specific versions: %v", deleteErr)
		}

		// Verify the current version is still 4
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to retrieve raw secret: %v", getRawErr)
		}

		if rawSecret.Metadata.CurrentVersion != 4 {
			t.Errorf(
				"Expected current version to remain 4, got %d",
				rawSecret.Metadata.CurrentVersion,
			)
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
			values, getErr := GetSecret(path, version)
			if getErr != nil {
				t.Errorf("Version %d should still be accessible: %v", version, getErr)
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

		path := "test/delete-all"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Delete all versions
		deleteErr := DeleteSecret(path, []int{1, 2, 3})
		if deleteErr != nil {
			t.Fatalf("Failed to delete all versions: %v", deleteErr)
		}

		// Verify the current version is 0 (no active versions)
		_, getRawErr := GetRawSecret(path, 0)
		if getRawErr == nil {
			t.Error(
				"Expected error when trying to get current version " +
					"of fully deleted secret",
			)
		}

		// Try to get the raw secret without version validation
		be := persist.Backend()
		// Do not pass a nil context even if the function allows it.
		secret, loadErr := be.LoadSecret(context.TODO(), path)
		if loadErr != nil {
			t.Fatalf("Failed to load raw secret: %v", loadErr)
		}

		if secret.Metadata.CurrentVersion != 0 {
			t.Errorf(
				"Expected current version to be 0, got %d",
				secret.Metadata.CurrentVersion,
			)
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

		err := DeleteSecret("test/nonexistent", []int{1})
		if err == nil {
			t.Error("Expected error when deleting non-existent secret")
		}
	})
}

func TestUndeleteSecret_SpecificVersions(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/undelete-specific"

		// Create and then delete multiple versions
		for i := 1; i <= 3; i++ {
			versionValues := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, versionValues)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Delete versions 1 and 2
		deleteErr := DeleteSecret(path, []int{1, 2})
		if deleteErr != nil {
			t.Fatalf("Failed to delete versions: %v", deleteErr)
		}

		// Undelete version 1
		undeleteErr := UndeleteSecret(path, []int{1})
		if undeleteErr != nil {
			t.Fatalf("Failed to undelete version 1: %v", undeleteErr)
		}

		// Verify version 1 is now accessible
		v1Values, getV1Err := GetSecret(path, 1)
		if getV1Err != nil {
			t.Errorf("Version 1 should be accessible after undelete: %v", getV1Err)
		}
		if v1Values["version"] != "v1" {
			t.Errorf("Expected version 1 to have value v1, got %s",
				v1Values["version"])
		}

		// Verify version 2 is still deleted
		_, getV2Err := GetSecret(path, 2)
		if getV2Err == nil {
			t.Error("Version 2 should still be deleted")
		}

		// Verify the current version is still 3
		currentValues, getCurrentErr := GetSecret(path, 0)
		if getCurrentErr != nil {
			t.Fatalf("Failed to get current version: %v", getCurrentErr)
		}
		if currentValues["version"] != "v3" {
			t.Errorf(
				"Expected current version to be v3, got %s",
				currentValues["version"],
			)
		}
	})
}

func TestUndeleteSecret_AllDeleted(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/undelete-all-deleted"

		// Create multiple versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Delete all versions
		deleteErr := DeleteSecret(path, []int{1, 2, 3})
		if deleteErr != nil {
			t.Fatalf("Failed to delete all versions: %v", deleteErr)
		}

		// Undelete without specifying versions (should undelete highest)
		undeleteErr := UndeleteSecret(path, []int{})
		if undeleteErr != nil {
			t.Fatalf("Failed to undelete: %v", undeleteErr)
		}

		// Verify version 3 is current and accessible
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to get raw secret: %v", getRawErr)
		}

		if rawSecret.Metadata.CurrentVersion != 3 {
			t.Errorf(
				"Expected current version to be 3, got %d",
				rawSecret.Metadata.CurrentVersion,
			)
		}

		values, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Errorf("Current version should be accessible: %v", getErr)
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

		path := "test/get-current"
		values := map[string]string{
			"key": "value",
		}

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret: %v", upsertErr)
		}

		// Get current version (version 0)
		retrievedValues, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to get current version: %v", getErr)
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

		path := "test/get-specific"

		// Create multiple versions
		expectedValues := make(map[int]map[string]string)
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
				"data":    fmt.Sprintf("data-%d", i),
			}
			expectedValues[i] = values

			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Retrieve each version specifically
		for version := 1; version <= 3; version++ {
			retrievedValues, getErr := GetSecret(path, version)
			if getErr != nil {
				t.Errorf("Failed to get version %d: %v", version, getErr)
				continue
			}

			if !reflect.DeepEqual(retrievedValues, expectedValues[version]) {
				t.Errorf(
					"Version %d: expected %v, got %v",
					version, expectedValues[version], retrievedValues,
				)
			}
		}
	})
}

func TestGetSecret_NonExistentSecret(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		_, err := GetSecret("test/nonexistent", 1)
		if err == nil {
			t.Error("Expected error when getting non-existent secret")
		}
	})
}

func TestGetSecret_NonExistentVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/nonexistent-version"
		values := map[string]string{
			"key": "value",
		}

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret: %v", upsertErr)
		}

		// Try to get a non-existent version
		_, getErr := GetSecret(path, 999)
		if getErr == nil {
			t.Error("Expected error when getting non-existent version")
		}
	})
}

func TestGetSecret_DeletedVersion(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/deleted-version"
		values := map[string]string{
			"key": "value",
		}

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret: %v", upsertErr)
		}

		// Delete version 1
		deleteErr := DeleteSecret(path, []int{1})
		if deleteErr != nil {
			t.Fatalf("Failed to delete version 1: %v", deleteErr)
		}

		// Try to get a deleted version
		_, getErr := GetSecret(path, 1)
		if getErr == nil {
			t.Error("Expected error when getting deleted version")
		}
	})
}

func TestGetRawSecret_WithMetadata(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "test/raw-secret"

		// Create multiple versions
		for i := 1; i <= 2; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
			}
			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Get raw secret
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to get raw secret: %v", getRawErr)
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

		path := "test/all-deleted"
		values := map[string]string{
			"key": "value",
		}

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret: %v", upsertErr)
		}

		// Delete all versions
		deleteErr := DeleteSecret(path, []int{1})
		if deleteErr != nil {
			t.Fatalf("Failed to delete version: %v", deleteErr)
		}

		// Try to get raw secret when all versions are deleted
		_, getRawErr := GetRawSecret(path, 0)
		if getRawErr == nil {
			t.Error("Expected error when all versions are deleted")
		}
	})
}

func TestSecretOperations_ConcurrentAccess(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		path := "/test/concurrent"

		// Create the initial secret
		upsertErr := UpsertSecret(path, map[string]string{"counter": "0"})
		if upsertErr != nil {
			t.Fatalf("Failed to create initial secret: %v", upsertErr)
		}

		// Test that multiple operations work correctly
		// Note: This is a simple test since the memory backend is not truly concurrent-safe,
		// But it tests the API works correctly in sequence

		operations := []func() *sdkErrors.SDKError{
			func() *sdkErrors.SDKError {
				return UpsertSecret(path, map[string]string{"counter": "1"})
			},
			func() *sdkErrors.SDKError {
				_, err := GetSecret(path, 0)
				return err
			},
			func() *sdkErrors.SDKError {
				return DeleteSecret(path, []int{1})
			},
			func() *sdkErrors.SDKError {
				return UndeleteSecret(path, []int{1})
			},
			func() *sdkErrors.SDKError {
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

		// Test with an empty map
		upsertEmptyErr := UpsertSecret(path, map[string]string{})
		if upsertEmptyErr != nil {
			t.Fatalf("Failed to create secret with empty values: %v", upsertEmptyErr)
		}

		// Retrieve and verify
		values, getEmptyErr := GetSecret(path, 0)
		if getEmptyErr != nil {
			t.Fatalf("Failed to get secret with empty values: %v", getEmptyErr)
		}

		if len(values) != 0 {
			t.Errorf("Expected empty values map, got %v", values)
		}

		// Test with a nil map
		upsertNilErr := UpsertSecret(path, nil)
		if upsertNilErr != nil {
			t.Fatalf("Failed to create secret with nil values: %v", upsertNilErr)
		}

		values, getNilErr := GetSecret(path, 0) // Should be version 2 now
		if getNilErr != nil {
			t.Fatalf("Failed to get secret with nil values: %v", getNilErr)
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

		// Create a large secret with many keys
		largeValues := make(map[string]string)
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%03d", i)
			value := fmt.Sprintf(
				"value_%03d_with_some_longer_content_to_make_it_bigger", i,
			)
			largeValues[key] = value
		}

		upsertErr := UpsertSecret(path, largeValues)
		if upsertErr != nil {
			t.Fatalf("Failed to create large secret: %v", upsertErr)
		}

		// Retrieve and verify
		retrievedValues, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to get large secret: %v", getErr)
		}

		if len(retrievedValues) != len(largeValues) {
			t.Errorf(
				"Expected %d keys, got %d",
				len(largeValues), len(retrievedValues),
			)
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
			"/test/with/deep/nested/pathPattern",
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

				upsertErr := UpsertSecret(path, values)
				if upsertErr != nil {
					t.Fatalf(
						"Failed to create secret with special characters: %v",
						upsertErr,
					)
				}

				retrievedValues, getErr := GetSecret(path, 0)
				if getErr != nil {
					t.Fatalf(
						"Failed to get secret with special characters: %v",
						getErr,
					)
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
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
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
		_ = UpsertSecret(path, values)
	}
}

func BenchmarkUpsertSecret_UpdateExisting(b *testing.B) {
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/update-secret"
	initialValues := map[string]string{
		"username": "admin",
		"password": "initial",
	}

	// Create the initial secret
	_ = UpsertSecret(path, initialValues)

	updatedValues := map[string]string{
		"username": "admin",
		"password": "updated",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updatedValues["counter"] = fmt.Sprintf("%d", i)
		_ = UpsertSecret(path, updatedValues)
	}
}

func BenchmarkGetSecret(b *testing.B) {
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/get-secret"
	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}

	_ = UpsertSecret(path, values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetSecret(path, 0)
	}
}

func BenchmarkGetRawSecret(b *testing.B) {
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	path := "/bench/get-raw-secret"
	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}

	_ = UpsertSecret(path, values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetRawSecret(path, 0)
	}
}
