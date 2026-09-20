//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// withEnvironment sets an environment variable for the remainder of the
// test and runs a function with it in place.
//
// t.Setenv restores the original value when the test finishes and fails the
// test if the variable cannot be set.
//
// Parameters:
//   - t: The test whose environment is modified.
//   - key: The environment variable to set.
//   - value: The value to assign to the variable.
//   - testFunc: The function to run with the variable in place.
func withEnvironment(t *testing.T, key, value string, testFunc func()) {
	t.Helper()
	t.Setenv(key, value)
	testFunc()
}

// unsetEnv removes an environment variable for the remainder of the test
// or benchmark and restores the original value afterwards.
//
// Unlike testing.TB.Setenv, it leaves the variable absent rather than empty,
// which matters for code that distinguishes "unset" from "".
//
// Parameters:
//   - tb: The test or benchmark whose environment is modified.
//   - key: The environment variable to remove.
func unsetEnv(tb testing.TB, key string) {
	tb.Helper()
	original, existed := os.LookupEnv(key)
	if unsetErr := os.Unsetenv(key); unsetErr != nil {
		tb.Fatalf("failed to unset %s: %v", key, unsetErr)
	}
	tb.Cleanup(func() {
		if !existed {
			return
		}
		if setErr := os.Setenv(key, original); setErr != nil {
			tb.Errorf("failed to restore %s: %v", key, setErr)
		}
	})
}

// closeBackend closes the active backend and fails the test if closing it
// returns an error.
//
// Parameters:
//   - t: The test to fail if Close reports an error.
//   - ctx: The context passed to the backend's Close method.
func closeBackend(t *testing.T, ctx context.Context) {
	t.Helper()
	if closeErr := persist.Backend().Close(ctx); closeErr != nil {
		t.Errorf("failed to close the backend: %v", closeErr)
	}
}

// Helper function to create a test key with a specific pattern
func createTestKeyWithPattern(pattern byte) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = pattern
	}
	return key
}

// Helper function to reset the root key to its zero state for tests
func resetRootKey() {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()
	for i := range rootKey {
		rootKey[i] = 0
	}
}

// Helper function to set the root key directly for testing (bypasses
// validation)
func setRootKeyDirect(key *[crypto.AES256KeySize]byte) {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()
	if key != nil {
		copy(rootKey[:], key[:])
	}
}

// Helper function to create a test key with random data
func createTestKey(t *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(key[:]); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	return key
}

// Helper function to create a test key with a specific pattern
func createPatternKey(pattern byte) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = pattern
	}
	return key
}

// createTestRootKey creates a deterministic root key for the SQLite backend
// tests.
//
// Parameters:
//   - _: The test, unused; kept so call sites read like the other key helpers.
//
// Returns:
//   - *[crypto.AES256KeySize]byte: A key whose bytes follow a fixed pattern.
func createTestRootKey(_ *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	// Use a predictable pattern for testing
	for i := range key {
		key[i] = byte(i % 256)
	}
	return key
}

// cleanupSQLiteDatabase removes the existing SQLite database so that the
// test or benchmark starts from a clean state.
//
// A database that cannot be removed would leak state into the run, so the
// failure is fatal.
//
// Parameters:
//   - tb: The test or benchmark to fail if the database cannot be removed.
func cleanupSQLiteDatabase(tb testing.TB) {
	tb.Helper()
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")

	if _, statErr := os.Stat(dbPath); statErr != nil {
		return
	}
	tb.Logf("Removing existing database at %s", dbPath)
	if rmErr := os.Remove(dbPath); rmErr != nil {
		tb.Fatalf("Failed to remove existing database: %v", rmErr)
	}
}

// withSQLiteEnvironment selects the SQLite backend with schema creation
// enabled for the remainder of the test and runs a function under it.
//
// Parameters:
//   - t: The test whose environment is modified.
//   - testFunc: The function to run with the environment in place.
func withSQLiteEnvironment(t *testing.T, testFunc func()) {
	t.Helper()
	t.Setenv(env.NexusBackendStore, "sqlite")
	unsetEnv(t, env.NexusDBSkipSchemaCreation)
	testFunc()
}

// getVersionNumbers lists the version numbers present in a secret.
//
// Parameters:
//   - secret: The secret whose versions are listed.
//
// Returns:
//   - []int: The version numbers, in map iteration order.
func getVersionNumbers(secret *kv.Value) []int {
	versions := make([]int, 0, len(secret.Versions))
	for v := range secret.Versions {
		versions = append(versions, v)
	}
	return versions
}

// resetBackendForTest is a placeholder for backend state isolation between
// tests; a fresh memory backend is initialized by each test instead.
func resetBackendForTest() {
}
