//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
	"github.com/spiffe/spike/internal/config"
)

func TestDataStore_loadSecretInternal_Success(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		defer func(store *DataStore, c context.Context) {
			_ = store.Close(c)
		}(store, context.Background())

		ctx := context.Background()
		path := "test/secret/path"

		// Test data
		versions := map[int]map[string]string{
			1: {
				"username": "admin",
				"password": "supersecret",
				"url":      "https://example.com",
			},
		}

		createdTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		metadata := TestSecretMetadata{
			CurrentVersion: 1,
			OldestVersion:  1,
			MaxVersions:    5,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
		}

		// Store test data directly in the database
		storeTestSecretDirectly(t, store, path, versions, metadata)

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results
		if err != nil {
			t.Errorf("loadSecretInternal failed: %v", err)
		}

		if secret == nil {
			t.Fatal("Expected non-nil secret")
		}

		// Check metadata
		if secret.Metadata.CurrentVersion != 1 {
			t.Errorf("Expected current version 1, got %d", secret.Metadata.CurrentVersion)
		}

		if secret.Metadata.OldestVersion != 1 {
			t.Errorf("Expected oldest version 1, got %d", secret.Metadata.OldestVersion)
		}

		if secret.Metadata.MaxVersions != 5 {
			t.Errorf("Expected max versions 5, got %d", secret.Metadata.MaxVersions)
		}

		if !secret.Metadata.CreatedTime.Equal(createdTime) {
			t.Errorf("Expected created time %v, got %v", createdTime, secret.Metadata.CreatedTime)
		}

		if !secret.Metadata.UpdatedTime.Equal(updatedTime) {
			t.Errorf("Expected updated time %v, got %v", updatedTime, secret.Metadata.UpdatedTime)
		}

		// Check versions
		if len(secret.Versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(secret.Versions))
		}

		version, exists := secret.Versions[1]
		if !exists {
			t.Fatal("Expected version 1 to exist")
		}

		// Check version data
		expectedData := versions[1]
		if len(version.Data) != len(expectedData) {
			t.Errorf("Expected %d data items, got %d", len(expectedData), len(version.Data))
		}

		for key, expectedValue := range expectedData {
			actualValue, exists := version.Data[key]
			if !exists {
				t.Errorf("Expected key '%s' to exist", key)
			}
			if actualValue != expectedValue {
				t.Errorf("Expected '%s'='%s', got '%s'", key, expectedValue, actualValue)
			}
		}

		if version.DeletedTime != nil {
			t.Error("Expected DeletedTime to be nil")
		}
	})
}

func TestDataStore_loadSecretInternal_MultipleVersions(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		defer func(store *DataStore, c context.Context) {
			_ = store.Close(c)
		}(store, context.Background())

		ctx := context.Background()
		path := "test/multi/versions"

		// Test data for different versions
		versions := map[int]map[string]string{
			1: {"key": "value1", "env": "dev"},
			2: {"key": "value2", "env": "prod"},
			3: {"key": "value3", "env": "prod", "region": "us-east-1"},
		}

		createdTime := time.Now().Add(-72 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		metadata := TestSecretMetadata{
			CurrentVersion: 3,
			OldestVersion:  1,
			MaxVersions:    5,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
		}

		// Store test data directly in the database
		storeTestSecretDirectly(t, store, path, versions, metadata)

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results
		if err != nil {
			t.Errorf("loadSecretInternal failed: %v", err)
		}

		if secret == nil {
			t.Fatal("Expected non-nil secret")
		}

		// Check metadata
		if secret.Metadata.CurrentVersion != 3 {
			t.Errorf("Expected current version 3, got %d", secret.Metadata.CurrentVersion)
		}

		// Check versions
		if len(secret.Versions) != 3 {
			t.Errorf("Expected 3 versions, got %d", len(secret.Versions))
		}

		// Verify version 1
		version1 := secret.Versions[1]
		if version1.Data["key"] != "value1" {
			t.Errorf("Expected v1 key='value1', got '%s'", version1.Data["key"])
		}
		if version1.DeletedTime != nil {
			t.Error("Expected v1 DeletedTime to be nil")
		}

		// Verify version 2 (should be marked as deleted by storeTestSecretDirectly)
		version2 := secret.Versions[2]
		if version2.Data["key"] != "value2" {
			t.Errorf("Expected v2 key='value2', got '%s'", version2.Data["key"])
		}
		if version2.DeletedTime == nil {
			t.Error("Expected v2 DeletedTime to be set")
		}

		// Verify version 3
		version3 := secret.Versions[3]
		if version3.Data["key"] != "value3" {
			t.Errorf("Expected v3 key='value3', got '%s'", version3.Data["key"])
		}
		if version3.Data["region"] != "us-east-1" {
			t.Errorf("Expected v3 region='us-east-1', got '%s'", version3.Data["region"])
		}
		if version3.DeletedTime != nil {
			t.Error("Expected v3 DeletedTime to be nil")
		}
	})
}

func TestDataStore_loadSecretInternal_SecretNotFound(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		defer func(store *DataStore, c context.Context) {
			_ = store.Close(c)
		}(store, context.Background())

		ctx := context.Background()
		path := "nonexistent/secret"

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results
		if err != nil {
			t.Errorf("loadSecretInternal should not return error for non-existent secret: %v", err)
		}

		if secret != nil {
			t.Error("Expected nil secret for non-existent path")
		}
	})
}

func TestDataStore_loadSecretInternal_EmptyVersionsResult(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		defer func(store *DataStore, c context.Context) {
			_ = store.Close(c)
		}(store, context.Background())

		ctx := context.Background()
		path := "test/empty/versions"

		// Insert metadata but no versions (inconsistent state)
		createdTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		_, err := store.db.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
			path, 1, 1, createdTime, updatedTime, 5)
		if err != nil {
			t.Fatalf("Failed to insert metadata: %v", err)
		}

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results - this should succeed but return a secret with no versions
		if err != nil {
			t.Errorf("loadSecretInternal should not fail with empty versions: %v", err)
		}

		if secret == nil {
			t.Fatal("Expected non-nil secret even with no versions")
		}

		// Check that metadata is preserved
		if secret.Metadata.CurrentVersion != 1 {
			t.Errorf("Expected current version 1, got %d", secret.Metadata.CurrentVersion)
		}

		// Check that the `versions` map is empty
		if len(secret.Versions) != 0 {
			t.Errorf("Expected no versions, got %d", len(secret.Versions))
		}
	})
}

func TestDataStore_loadSecretInternal_CorruptedData(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		defer func(store *DataStore, c context.Context) {
			_ = store.Close(c)
		}(store, context.Background())

		ctx := context.Background()
		path := "test/corrupted/data"

		createdTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		// Insert metadata
		_, err := store.db.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
			path, 1, 1, createdTime, updatedTime, 5)
		if err != nil {
			t.Fatalf("Failed to insert metadata: %v", err)
		}

		// Insert corrupted version data (invalid nonce/encrypted combination)
		invalidNonce := make([]byte, store.Cipher.NonceSize())
		invalidEncrypted := []byte("invalid encrypted data that will fail decryption")

		versionCreatedTime := createdTime.Add(1 * time.Hour)
		_, err = store.db.ExecContext(ctx, ddl.QueryUpsertSecret,
			path, 1, invalidNonce, invalidEncrypted, versionCreatedTime, nil)
		if err != nil {
			t.Fatalf("Failed to insert corrupted version: %v", err)
		}

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results - should fail due to the decryption error
		if err == nil {
			t.Error("Expected error for corrupted encrypted data")
		}

		if secret != nil {
			t.Error("Expected nil secret on decryption error")
		}

		if err != nil && !contains(err.Error(), "failed to decrypt secret version") {
			t.Errorf("Expected decryption error, got: %v", err)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark test for performance with real SQLite
func BenchmarkDataStore_loadSecretInternal(b *testing.B) {
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
	dataDir := config.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = os.Remove(dbPath)
	}

	store := createTestDataStore(b)
	defer func(store *DataStore, c context.Context) {
		_ = store.Close(c)
	}(store, context.Background())

	ctx := context.Background()
	path := "benchmark/secret"

	// Setup test data
	versions := map[int]map[string]string{
		1: {
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	metadata := TestSecretMetadata{
		CurrentVersion: 1,
		OldestVersion:  1,
		MaxVersions:    5,
		CreatedTime:    time.Now().Add(-24 * time.Hour).Truncate(time.Second),
		UpdatedTime:    time.Now().Add(-1 * time.Hour).Truncate(time.Second),
	}

	storeTestSecretDirectly(b, store, path, versions, metadata)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.loadSecretInternal(ctx, path)
		if err != nil {
			b.Errorf("loadSecretInternal failed: %v", err)
		}
	}
}
