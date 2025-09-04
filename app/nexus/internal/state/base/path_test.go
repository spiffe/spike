//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestListKeys_EmptyBackend(t *testing.T) {
	// Test with a memory backend that has no secrets
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		// Reset and initialize with a memory backend
		resetBackendForTest()
		persist.InitializeBackend(nil)

		keys := ListKeys()

		if len(keys) != 0 {
			t.Errorf("Expected empty slice, got %d keys: %v", len(keys), keys)
		}
	})
}

// Helper function to reset the backend state for testing
func resetBackendForTest() {
	// This will be implemented to ensure a clean state between tests
	// For now, we rely on initializing a fresh memory backend
}

func TestListKeys_SingleSecret(t *testing.T) {
	// This test will use the actual memory backend
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		// Since we can't easily mock the backend, we'll store a secret first
		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		err := backend.StoreSecret(ctx, "test/secret1", testSecret)
		if err != nil {
			t.Fatalf("Failed to store test secret: %v", err)
		}

		keys := ListKeys()

		expected := []string{"test/secret1"}
		if !reflect.DeepEqual(keys, expected) {
			t.Errorf("Expected keys %v, got %v", expected, keys)
		}
	})
}

func TestListKeys_MultipleSecrets(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Store multiple secrets with different paths
		testPaths := []string{
			"app/db/password",
			"app/api/key",
			"service/token",
			"admin/credentials",
		}

		for _, path := range testPaths {
			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret at %s: %v", path, err)
			}
		}

		keys := ListKeys()

		// Keys should be returned sorted
		expectedSorted := make([]string, len(testPaths))
		copy(expectedSorted, testPaths)
		sort.Strings(expectedSorted)

		if !reflect.DeepEqual(keys, expectedSorted) {
			t.Errorf("Expected keys %v, got %v", expectedSorted, keys)
		}
	})
}

func TestListKeys_SortingBehavior(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Store secrets in non-alphabetical order
		unsortedPaths := []string{
			"zebra",
			"apple",
			"banana",
			"cherry",
			"dog",
			"ant",
		}

		for _, path := range unsortedPaths {
			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret at %s: %v", path, err)
			}
		}

		keys := ListKeys()

		// Verify they come back sorted
		expected := []string{"ant", "apple", "banana", "cherry", "dog", "zebra"}

		if !reflect.DeepEqual(keys, expected) {
			t.Errorf("Expected sorted keys %v, got %v", expected, keys)
		}

		// Verify they are indeed sorted
		if !sort.StringsAreSorted(keys) {
			t.Error("Keys should be sorted in lexicographical order")
		}
	})
}

func TestListKeys_SpecialCharacters(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Test paths with special characters
		specialPaths := []string{
			"app/db-prod/password",
			"app/db_test/password",
			"service/api.key",
			"service/config/app-name",
			"service/config/app_name",
		}

		for _, path := range specialPaths {
			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret at %s: %v", path, err)
			}
		}

		keys := ListKeys()

		// Verify all paths are returned
		if len(keys) != len(specialPaths) {
			t.Errorf("Expected %d keys, got %d", len(specialPaths), len(keys))
		}

		// Verify sorting works correctly with special characters
		if !sort.StringsAreSorted(keys) {
			t.Error("Keys with special characters should be sorted correctly")
		}

		// Check specific ordering expectations
		for _, expectedPath := range specialPaths {
			found := false
			for _, actualPath := range keys {
				if actualPath == expectedPath {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected path %s not found in result", expectedPath)
			}
		}
	})
}

func TestListKeys_DeepPaths(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Test deeply nested paths
		deepPaths := []string{
			"/",
			"app",
			"app/service",
			"app/service/db",
			"app/service/db/prod",
			"app/service/db/prod/password",
			"app/service/api/v1/key",
			"app/service/cache/redis/config",
		}

		for _, path := range deepPaths {
			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret at %s: %v", path, err)
			}
		}

		keys := ListKeys()

		// Verify all paths are returned and sorted
		if len(keys) != len(deepPaths) {
			t.Errorf("Expected %d keys, got %d", len(deepPaths), len(keys))
		}

		if !sort.StringsAreSorted(keys) {
			t.Error("Deep paths should be sorted correctly")
		}
	})
}

func TestListKeys_DuplicatePaths(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		secret1 := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"version": "1"},
					Version: 1,
				},
			},
		}

		secret2 := kv.Value{
			Versions: map[int]kv.Version{
				2: {
					Data:    map[string]string{"version": "2"},
					Version: 2,
				},
			},
		}

		path := "test/duplicate"

		// Store first version
		err := backend.StoreSecret(ctx, path, secret1)
		if err != nil {
			t.Fatalf("Failed to store first secret: %v", err)
		}

		// Overwrite with the second version
		err = backend.StoreSecret(ctx, path, secret2)
		if err != nil {
			t.Fatalf("Failed to store second secret: %v", err)
		}

		keys := ListKeys()

		// Should only appear once in the list
		expectedCount := 1
		actualCount := 0
		for _, key := range keys {
			if key == path {
				actualCount++
			}
		}

		if actualCount != expectedCount {
			t.Errorf("Expected path %s to appear %d time(s), got %d", path, expectedCount, actualCount)
		}
	})
}

func TestListKeys_LargeBatch(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Generate a large number of paths
		numSecrets := 100
		expectedPaths := make([]string, numSecrets)

		for i := 0; i < numSecrets; i++ {
			path := fmt.Sprintf("batch/secret-%04d", i)
			expectedPaths[i] = path

			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret %d: %v", i, err)
			}
		}

		keys := ListKeys()

		// Verify count
		if len(keys) != numSecrets {
			t.Errorf("Expected %d keys, got %d", numSecrets, len(keys))
		}

		// Verify sorting
		if !sort.StringsAreSorted(keys) {
			t.Error("Large batch of keys should be sorted")
		}

		// Verify all expected paths are present
		sort.Strings(expectedPaths) // Sort expected for comparison
		if !reflect.DeepEqual(keys, expectedPaths) {
			t.Error("Returned keys don't match expected paths")
		}
	})
}

func TestListKeys_MemoryReuse(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		resetBackendForTest()
		persist.InitializeBackend(nil)

		backend := persist.Backend()
		ctx := context.Background()

		testSecret := kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:    map[string]string{"key": "value"},
					Version: 1,
				},
			},
		}

		// Store some secrets
		paths := []string{"test1", "test2", "test3"}
		for _, path := range paths {
			err := backend.StoreSecret(ctx, path, testSecret)
			if err != nil {
				t.Fatalf("Failed to store secret: %v", err)
			}
		}

		// Call ListKeys multiple times to ensure no memory issues
		for i := 0; i < 10; i++ {
			keys := ListKeys()
			if len(keys) != 3 {
				t.Errorf("Iteration %d: expected 3 keys, got %d", i, len(keys))
			}
		}
	})
}

// Benchmark tests
func BenchmarkListKeys_Empty(b *testing.B) {
	// Save and restore environment variable
	original := os.Getenv(env.NexusBackendStore)
	_ = os.Setenv(env.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(env.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListKeys()
	}
}

func BenchmarkListKeys_SmallSet(b *testing.B) {
	original := os.Getenv(env.NexusBackendStore)
	_ = os.Setenv(env.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(env.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)
	backend := persist.Backend()
	ctx := context.Background()

	testSecret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"key": "value"},
				Version: 1,
			},
		},
	}

	// Set up a small set of secrets
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/bench/secret-%d", i)
		_ = backend.StoreSecret(ctx, path, testSecret)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListKeys()
	}
}

func BenchmarkListKeys_LargeSet(b *testing.B) {
	original := os.Getenv(env.NexusBackendStore)
	_ = os.Setenv(env.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(env.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	resetBackendForTest()
	persist.InitializeBackend(nil)
	backend := persist.Backend()
	ctx := context.Background()

	testSecret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"key": "value"},
				Version: 1,
			},
		},
	}

	// Set up a large set of secrets
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("bench/large/secret-%04d", i)
		_ = backend.StoreSecret(ctx, path, testSecret)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListKeys()
	}
}
