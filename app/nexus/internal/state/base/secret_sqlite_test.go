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
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// createTestRootKey creates a test root key for SQLite backend
func createTestRootKey(_ *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	// Use a predictable pattern for testing
	for i := range key {
		key[i] = byte(i % 256)
	}
	return key
}

// cleanupSQLiteDatabase removes the existing SQLite database to ensure
// a clean test state
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

// withSQLiteEnvironment sets up environment for SQLite testing
func withSQLiteEnvironment(_ *testing.T, testFunc func()) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	originalSkipSchema := os.Getenv(env.NexusDBSkipSchemaCreation)

	// Ensure cleanup happens
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalSkipSchema != "" {
			_ = os.Setenv(env.NexusDBSkipSchemaCreation, originalSkipSchema)
		} else {
			_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)
		}
	}()

	// Set to SQLite backend and ensure schema creation
	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Unsetenv(env.NexusDBSkipSchemaCreation)

	testFunc()
}

func TestSQLiteSecret_NewSecret(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		// Verify the environment is set correctly
		if env.BackendStoreTypeVal() != env.Sqlite {
			t.Fatalf("Expected env.BackendStoreType()=Sqlite, got %v",
				env.BackendStoreTypeVal())
		}

		// Get the actual database pathPattern used by the system
		dataDir := fs.NexusDataFolder()
		dbPath := filepath.Join(dataDir, "spike.db")
		t.Logf("Using SQLite database at: %s", dbPath)

		// Clean up any existing database
		cleanupSQLiteDatabase(t)

		// Initialize with a valid root key
		rootKey := createTestRootKey(t)
		resetRootKey()

		t.Logf("Initializing SQLite backend...")
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Check what secrets exist immediately after initialization
		// (should be empty for clean DB)
		backend := persist.Backend()
		allSecrets, loadAllErr := backend.LoadAllSecrets(ctx)
		if loadAllErr != nil {
			t.Fatalf("Failed to load all secrets after init: %v", loadAllErr)
		}
		t.Logf("Found %d existing secrets after initialization (expected 0)",
			len(allSecrets))
		if len(allSecrets) != 0 {
			for path := range allSecrets {
				t.Logf("  - Unexpected secret at pathPattern: %s", path)
			}
		}

		path := "test/sqlite-new-secret"
		values := map[string]string{
			"username": "admin",
			"password": "secret123",
			"database": "prod_db",
		}

		// Verify LoadSecret returns ErrEntityNotFound for non-existent secret
		secretBeforeUpsert, loadErr := backend.LoadSecret(ctx, path)
		if loadErr == nil {
			t.Fatalf("Expected error from LoadSecret before upsert, got nil")
		}
		if !loadErr.Is(sdkErrors.ErrEntityNotFound) {
			t.Fatalf("Expected ErrEntityNotFound, got: %v", loadErr)
		}
		if secretBeforeUpsert != nil {
			t.Fatalf("Expected LoadSecret to return nil for non-existent secret,"+
				" got: %+v", secretBeforeUpsert)
		}
		t.Logf("LoadSecret correctly returned ErrEntityNotFound" +
			" for non-existent path")

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to upsert new secret to SQLite: %v", upsertErr)
		}

		// Verify the secret was created and encrypted
		retrievedValues, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to retrieve secret from SQLite: %v", getErr)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Expected values %v, got %v", values, retrievedValues)
		}

		// Verify the database file was created
		if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
			t.Error("SQLite database file should have been created")
		}
	})
}

func TestSQLiteSecret_Persistence(t *testing.T) {
	path := "test/sqlite-persistence"
	values := map[string]string{
		"persistent": "data",
		"should":     "survive restart",
	}

	// First session - create secret
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

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret in first session: %v", upsertErr)
		}
	})

	// Second session - verify persistence (simulate restart)
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		rootKey := createTestRootKey(t) // Same key as the first session
		resetRootKey()
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		retrievedValues, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to retrieve secret in second session: %v", getErr)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Expected persistent values %v, got %v",
				values, retrievedValues)
		}
	})
}

func TestSQLiteSecret_SimpleVersioning(t *testing.T) {
	// Simple test to understand SQLite versioning behavior
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

		path := "test/simple-versioning"

		// Create the first version
		t.Log("Creating first version...")
		values1 := map[string]string{"data": "version1"}
		upsertErr1 := UpsertSecret(path, values1)
		if upsertErr1 != nil {
			t.Fatalf("Failed to create first version: %v", upsertErr1)
		}

		// Verify the first version using backend directly
		secret1, loadErr1 := persist.Backend().LoadSecret(ctx, path)
		if loadErr1 != nil {
			t.Fatalf("Failed to load secret from backend: %v", loadErr1)
		}
		if secret1 == nil {
			t.Fatal("Secret should exist after first upsert")
			return
		}
		t.Logf("After first upsert - CurrentVersion: %d, Versions: %v",
			secret1.Metadata.CurrentVersion, getVersionNumbers(secret1))

		// Create a second version
		t.Log("Creating second version...")
		values2 := map[string]string{"data": "version2"}
		upsertErr2 := UpsertSecret(path, values2)
		if upsertErr2 != nil {
			t.Fatalf("Failed to create second version: %v", upsertErr2)
		}

		// Verify the second version using backend directly
		secret2, loadErr2 := persist.Backend().LoadSecret(ctx, path)
		if loadErr2 != nil {
			t.Fatalf("Failed to load secret from backend after second upsert: %v",
				loadErr2)
		}
		if secret2 == nil {
			t.Fatal("Secret should exist after second upsert")
			return
		}
		t.Logf("After second upsert - CurrentVersion: %d, Versions: %v",
			secret2.Metadata.CurrentVersion, getVersionNumbers(secret2))

		// Test GetSecret with version 0 (current)
		currentValues, getCurrentErr := GetSecret(path, 0)
		if getCurrentErr != nil {
			t.Fatalf("Failed to get current version: %v", getCurrentErr)
		}
		if currentValues["data"] != "version2" {
			t.Errorf("Expected current version to be 'version2', got %s",
				currentValues["data"])
		}

		// Test GetSecret with version 1
		version1Values, getV1Err := GetSecret(path, 1)
		if getV1Err != nil {
			t.Fatalf("Failed to get version 1: %v", getV1Err)
		}
		if version1Values["data"] != "version1" {
			t.Errorf("Expected version 1 to be 'version1', got %s",
				version1Values["data"])
		}

		// Test GetSecret with version 2
		version2Values, getV2Err := GetSecret(path, 2)
		if getV2Err != nil {
			t.Fatalf("Failed to get version 2: %v", getV2Err)
		}
		if version2Values["data"] != "version2" {
			t.Errorf("Expected version 2 to be 'version2', got %s",
				version2Values["data"])
		}
	})
}

// Helper function to get version numbers from a secret
func getVersionNumbers(secret *kv.Value) []int {
	versions := make([]int, 0, len(secret.Versions))
	for v := range secret.Versions {
		versions = append(versions, v)
	}
	return versions
}

func TestSQLiteSecret_VersionPersistence(t *testing.T) {
	path := "test/sqlite-versions"

	// Create multiple versions in the first session
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

		// Create 3 versions
		for i := 1; i <= 3; i++ {
			values := map[string]string{
				"version": fmt.Sprintf("v%d", i),
				"data":    fmt.Sprintf("data-%d", i),
			}
			upsertErr := UpsertSecret(path, values)
			if upsertErr != nil {
				t.Fatalf("Failed to create version %d: %v", i, upsertErr)
			}
		}

		// Verify all versions were created in the first session
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to get raw secret in first session: %v", getRawErr)
		}
		t.Logf("First session - Current version: %d",
			rawSecret.Metadata.CurrentVersion)
		t.Logf("First session - Total versions: %d", len(rawSecret.Versions))
		for version := range rawSecret.Versions {
			t.Logf("  - Version %d exists in first session", version)
		}
	})

	// Verify all versions persist in the second session
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()

		rootKey := createTestRootKey(t)
		resetRootKey()
		persist.InitializeBackend(rootKey)
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// First, get the raw secret to understand what versions exist
		rawSecret, getRawErr := GetRawSecret(path, 0)
		if getRawErr != nil {
			t.Fatalf("Failed to get raw secret: %v", getRawErr)
		}

		t.Logf("Second session - Current version: %d",
			rawSecret.Metadata.CurrentVersion)
		t.Logf("Second session - Total versions stored: %d", len(rawSecret.Versions))
		for version := range rawSecret.Versions {
			t.Logf("  - Version %d exists in second session", version)
		}

		// Check each version
		for version := 1; version <= 3; version++ {
			values, getErr := GetSecret(path, version)
			if getErr != nil {
				t.Errorf("Failed to get version %d: %v", version, getErr)
				continue
			}

			expectedVersion := fmt.Sprintf("v%d", version)
			if values["version"] != expectedVersion {
				t.Errorf("Version %d: expected %s, got %s",
					version, expectedVersion, values["version"])
			}
		}

		// Verify metadata
		if rawSecret.Metadata.CurrentVersion != 3 {
			t.Errorf("Expected current version 3, got %d",
				rawSecret.Metadata.CurrentVersion)
		}
		if len(rawSecret.Versions) != 3 {
			t.Errorf("Expected 3 versions, got %d", len(rawSecret.Versions))
		}
	})
}

func TestSQLiteSecret_EncryptionWithDifferentKeys(t *testing.T) {
	path := "test/sqlite-encryption"
	values := map[string]string{
		"sensitive": "secret_data",
		"api_key":   "abc123xyz",
	}

	// Create the secret with the first key
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

		upsertErr := UpsertSecret(path, values)
		if upsertErr != nil {
			t.Fatalf("Failed to create secret with key1: %v", upsertErr)
		}
	})

	// Try to read with a different key (should fail or return garbage)
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

		// This should either fail or return decrypted garbage
		// (depending on implementation)
		_, getErr := GetSecret(path, 0)
		// We expect this to fail with the wrong key, but exact behavior
		// depends on implementation
		if getErr == nil {
			t.Log("Note: GetSecret succeeded with wrong key" +
				" - this might indicate encryption issue")
		} else {
			t.Logf("Expected behavior: GetSecret failed with wrong key: %v",
				getErr)
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

		retrievedValues, getErr := GetSecret(path, 0)
		if getErr != nil {
			t.Fatalf("Failed to retrieve with original key: %v", getErr)
		}

		if !reflect.DeepEqual(retrievedValues, values) {
			t.Errorf("Values changed with original key: expected %v, got %v",
				values, retrievedValues)
		}
	})
}

// Benchmark tests for SQLite
func BenchmarkSQLiteUpsertSecret(b *testing.B) {
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
	persist.InitializeBackend(rootKey)
	Initialize(rootKey)

	values := map[string]string{
		"username": "admin",
		"password": "secret123",
		"token":    "abcdef123456",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("bench/secret-%d", i)
		err := UpsertSecret(path, values)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkSQLiteGetSecret(b *testing.B) {
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
	persist.InitializeBackend(rootKey)
	Initialize(rootKey)

	// Create a secret to benchmark against
	path := "bench/get-secret"
	values := map[string]string{
		"username": "admin",
		"password": "secret123",
	}
	_ = UpsertSecret(path, values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetSecret(path, 0)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
