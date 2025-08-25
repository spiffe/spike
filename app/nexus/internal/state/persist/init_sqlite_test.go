//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike/internal/config"
)

// cleanupSQLiteDatabase removes the existing SQLite database to ensure clean test state
func cleanupSQLiteDatabase(t *testing.T) {
	dataDir := config.SpikeNexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")

	// Remove the database file if it exists
	if _, err := os.Stat(dbPath); err == nil {
		t.Logf("Removing existing database at %s", dbPath)
		if err := os.Remove(dbPath); err != nil {
			t.Logf("Warning: Failed to remove existing database: %v", err)
		}
	}
}

func TestInitializeSqliteBackend_Success(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create test key
	key := createTestKey(t)

	// Test initialization
	backend := initializeSqliteBackend(key)

	// Verify backend was created successfully
	if backend == nil {
		t.Fatal("initializeSqliteBackend returned nil")
	}

	// Verify database file was created
	dataDir := config.SpikeNexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestInitializeSqliteBackend_NilKey(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Test with nil key - this should panic due to nil pointer dereference
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when initializing sqlite backend with nil key")
		}
	}()

	initializeSqliteBackend(nil)
}

func TestInitializeSqliteBackend_ZeroKey(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create zero key
	zeroKey := createZeroKey()

	// Test with zero key - this should work (unlike the general validation in InitializeBackend)
	backend := initializeSqliteBackend(zeroKey)

	// Should succeed - zero key is valid for SQLite backend creation itself
	if backend == nil {
		t.Error("Expected initializeSqliteBackend to succeed with zero key")
	}
}

func TestInitializeSqliteBackend_MultipleInitializations(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create test key
	key := createTestKey(t)

	// Initialize first time
	backend1 := initializeSqliteBackend(key)
	if backend1 == nil {
		t.Fatal("First initialization failed")
	}

	// Initialize second time (should also work)
	backend2 := initializeSqliteBackend(key)
	if backend2 == nil {
		t.Fatal("Second initialization failed")
	}

	// Both should be valid but different instances
	if backend1 == backend2 {
		t.Error("Expected different backend instances")
	}
}

// Benchmark tests
func BenchmarkInitializeSqliteBackend(b *testing.B) {
	// Create test key
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = byte(i % 256)
	}

	// Clean up before and after
	dataDir := config.SpikeNexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		os.Remove(dbPath)
	}
	defer func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend := initializeSqliteBackend(key)
		if backend == nil {
			b.Fatal("initializeSqliteBackend returned nil")
		}
	}
}
