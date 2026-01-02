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
)

func TestInitializeSqliteBackend_Success(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create the test key
	key := createTestKey(t)

	// Test initialization
	backend := initializeSqliteBackend(key)

	// Verify backend was created successfully
	if backend == nil {
		t.Fatal("initializeSqliteBackend returned nil")
	}

	// Verify the database file was created
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestInitializeSqliteBackend_NilKey(t *testing.T) {
	// Enable panic mode for log.FatalErr so we can test error handling
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")

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
	// Enable panic mode for log.FatalErr so we can test error handling
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")

	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create a zero key
	zeroKey := createZeroKey()

	// Test with a zero key - this should panic because you cannot encrypt
	// a database with an empty key.
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when initializing sqlite backend with zero key")
		}
	}()

	initializeSqliteBackend(zeroKey)

	t.Error("Should have panicked")
}

func TestInitializeSqliteBackend_MultipleInitializations(t *testing.T) {
	cleanupSQLiteDatabase(t)
	defer cleanupSQLiteDatabase(t)

	// Create a test key
	key := createTestKey(t)

	// Initialize first time
	backend1 := initializeSqliteBackend(key)
	if backend1 == nil {
		t.Fatal("First initialization failed")
	}

	// Initialize the second time (should also work)
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
	// Create a test key
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = byte(i % 256)
	}

	// Clean up before and after
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}
	defer func() {
		if _, err := os.Stat(dbPath); err == nil {
			_ = os.Remove(dbPath)
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
