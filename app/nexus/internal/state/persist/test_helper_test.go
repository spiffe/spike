//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
)

// cleanupSQLiteDatabase removes the existing SQLite database to ensure a clean
// test state
func cleanupSQLiteDatabase(t *testing.T) {
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")

	// Remove the database file if it exists
	if _, err := os.Stat(dbPath); err == nil {
		t.Logf("Removing existing database at %s", dbPath)
		if err := os.Remove(dbPath); err != nil {
			t.Logf("Warning: Failed to remove existing database: %v", err)
		}
	}
}

// Helper function to create a test root key with a specific pattern
func createTestKey(_ *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = byte(i % 256) // Predictable pattern for testing
	}
	return key
}

// Helper function to create a zero key
func createZeroKey() *[crypto.AES256KeySize]byte {
	return &[crypto.AES256KeySize]byte{} // All zeros
}

// withEnvironment runs a function with an environment variable set.
//
// It works for tests and benchmarks alike. tb.Setenv restores the previous
// value when the test ends; an empty value is what the SDK treats as
// "unset".
//
// Parameters:
//   - tb: The test or benchmark whose environment is modified.
//   - key: The environment variable to set.
//   - value: The value to assign to the variable.
//   - testFunc: The function to run with the variable in place.
func withEnvironment(tb testing.TB, key, value string, testFunc func()) {
	tb.Helper()
	tb.Setenv(key, value)

	testFunc()
}

// removeTempDir deletes the per-run data directory.
//
// A failure here does not change the test verdict, but it must not pass
// silently either, so it is logged as a warning.
//
// Parameters:
//   - dir: The temporary data directory to remove.
func removeTempDir(dir string) {
	if rmErr := os.RemoveAll(dir); rmErr != nil {
		log.Warn("failed to remove the temporary data directory",
			"dir", dir, "err", rmErr)
	}
}
