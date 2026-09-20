//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

func TestDataStore_loadSecretInternal_Success(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		closeStoreOnCleanup(t, store)

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
			t.Fatalf("loadSecretInternal failed: %v", err)
		}

		if secret == nil {
			t.Fatal("Expected non-nil secret")
			return
		}

		// Check metadata
		if secret.Metadata.CurrentVersion != 1 {
			t.Errorf(
				"Expected current version 1, got %d", secret.Metadata.CurrentVersion,
			)
		}

		if secret.Metadata.OldestVersion != 1 {
			t.Errorf(
				"Expected oldest version 1, got %d", secret.Metadata.OldestVersion,
			)
		}

		if secret.Metadata.MaxVersions != 5 {
			t.Errorf("Expected max versions 5, got %d", secret.Metadata.MaxVersions)
		}

		if !secret.Metadata.CreatedTime.Equal(createdTime) {
			t.Errorf(
				"Expected created time %v, got %v",
				createdTime,
				secret.Metadata.CreatedTime,
			)
		}

		if !secret.Metadata.UpdatedTime.Equal(updatedTime) {
			t.Errorf(
				"Expected updated time %v, got %v",
				updatedTime,
				secret.Metadata.UpdatedTime,
			)
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
			t.Errorf(
				"Expected %d data items, got %d", len(expectedData), len(version.Data),
			)
		}

		for key, expectedValue := range expectedData {
			actualValue, exists := version.Data[key]
			if !exists {
				t.Errorf("Expected key '%s' to exist", key)
			}
			if actualValue != expectedValue {
				t.Errorf(
					"Expected '%s'='%s', got '%s'", key, expectedValue, actualValue,
				)
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
		closeStoreOnCleanup(t, store)

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
			t.Fatalf("loadSecretInternal failed: %v", err)
		}

		if secret == nil {
			t.Fatal("Expected non-nil secret")
			return
		}

		// Check metadata
		if secret.Metadata.CurrentVersion != 3 {
			t.Errorf(
				"Expected current version 3, got %d", secret.Metadata.CurrentVersion,
			)
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
			t.Errorf(
				"Expected v3 region='us-east-1', got '%s'", version3.Data["region"],
			)
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
		closeStoreOnCleanup(t, store)

		ctx := context.Background()
		path := "nonexistent/secret"

		// Execute the function
		secret, err := store.loadSecretInternal(ctx, path)

		// Verify results: should return ErrEntityNotFound for the non-existent
		// secret
		if err == nil {
			t.Error("Expected ErrEntityNotFound for non-existent secret")
		}

		if err != nil && !err.Is(sdkErrors.ErrEntityNotFound) {
			t.Errorf("Expected ErrEntityNotFound, got: %v", err)
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
		closeStoreOnCleanup(t, store)

		ctx := context.Background()
		path := "test/empty/versions"

		// Insert metadata but no versions (inconsistent state).
		// CurrentVersion=1 but no actual version 1 exists.
		createdTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		metadata := TestSecretMetadata{
			CurrentVersion: 1,
			OldestVersion:  1,
			MaxVersions:    5,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
		}
		insertEncryptedMetadata(ctx, t, store, path, metadata)

		// Execute the function
		secret, loadErr := store.loadSecretInternal(ctx, path)

		// Verify results: should fail with ErrStateIntegrityCheck because
		// CurrentVersion=1 but version 1 doesn't exist in the Versions map.
		if loadErr == nil {
			t.Error("Expected ErrStateIntegrityCheck for inconsistent state")
		}

		if loadErr != nil && !loadErr.Is(sdkErrors.ErrStateIntegrityCheck) {
			t.Errorf("Expected ErrStateIntegrityCheck, got: %v", loadErr)
		}

		if secret != nil {
			t.Error("Expected nil secret for integrity check failure")
		}
	})
}

func TestDataStore_loadSecretInternal_CorruptedData(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		closeStoreOnCleanup(t, store)

		ctx := context.Background()
		path := "test/corrupted/data"

		createdTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		updatedTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		metadata := TestSecretMetadata{
			CurrentVersion: 1,
			OldestVersion:  1,
			MaxVersions:    5,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
		}
		// Insert metadata
		insertEncryptedMetadata(ctx, t, store, path, metadata)

		// Insert corrupted version data (invalid nonce/encrypted combination)
		invalidNonce := make([]byte, store.Cipher.NonceSize())
		invalidEncrypted := []byte(
			"invalid encrypted data that will fail decryption",
		)

		versionCreatedTime := createdTime.Add(1 * time.Hour)
		_, err := store.db.ExecContext(ctx, ddl.QueryUpsertVersion,
			path, 1, invalidNonce, invalidEncrypted, versionCreatedTime, nil)
		if err != nil {
			t.Fatalf("Failed to insert corrupted version: %v", err)
		}

		// Execute the function
		secret, loadErr := store.loadSecretInternal(ctx, path)

		// Verify results - should fail with ErrCryptoDecryptionFailed
		if loadErr == nil {
			t.Error("Expected error for corrupted encrypted data")
		}

		if secret != nil {
			t.Error("Expected nil secret on decryption error")
		}

		if loadErr != nil && !loadErr.Is(sdkErrors.ErrCryptoDecryptionFailed) {
			t.Errorf("Expected ErrCryptoDecryptionFailed, got: %v", loadErr)
		}
	})
}

// TestDataStore_loadSecretInternal_CorruptMetadata verifies that a metadata
// row whose decrypted fields do not parse as integers is rejected with an
// integrity error that names the field, and that no secret is returned.
func TestDataStore_loadSecretInternal_CorruptMetadata(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		cleanupSQLiteDatabase(t)
		store := createTestDataStore(t)
		closeStoreOnCleanup(t, store)

		ctx := context.Background()
		path := "test/corrupt/metadata"

		now := time.Now().Truncate(time.Second)
		healthy := rawSecretMetadata{
			currentVersion: "1",
			oldestVersion:  "1",
			maxVersions:    "5",
			createdTime:    strconv.FormatInt(now.Unix(), 10),
			updatedTime:    strconv.FormatInt(now.Unix(), 10),
		}

		// Each case corrupts exactly one field. The loader must refuse the
		// row instead of silently reading the field as zero.
		cases := []struct {
			name   string
			mutate func(raw *rawSecretMetadata)
		}{
			{"current_version", func(r *rawSecretMetadata) {
				r.currentVersion = "not-a-number"
			}},
			{"oldest_version", func(r *rawSecretMetadata) {
				r.oldestVersion = ""
			}},
			{"created_time", func(r *rawSecretMetadata) {
				r.createdTime = "2025-01-01"
			}},
			{"updated_time", func(r *rawSecretMetadata) {
				r.updatedTime = "1.5"
			}},
			{"max_versions", func(r *rawSecretMetadata) {
				r.maxVersions = "five"
			}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				raw := healthy
				tc.mutate(&raw)
				insertRawEncryptedMetadata(ctx, t, store, path, raw)

				secret, loadErr := store.loadSecretInternal(ctx, path)
				if loadErr == nil {
					t.Fatalf("expected an error for a corrupt %s", tc.name)
					return
				}
				if !loadErr.Is(sdkErrors.ErrStateIntegrityCheck) {
					t.Errorf("expected ErrStateIntegrityCheck, got: %v", loadErr)
				}
				if !strings.Contains(loadErr.Msg, tc.name) {
					t.Errorf("expected the message to name %s, got: %q",
						tc.name, loadErr.Msg)
				}
				if secret != nil {
					t.Error("expected a nil secret for corrupt metadata")
				}
			})
		}
	})
}

// Benchmark test for performance with real SQLite
func BenchmarkDataStore_loadSecretInternal(b *testing.B) {
	// Select the SQLite backend with schema creation enabled. b.Setenv
	// restores the previous values when the benchmark ends; an empty
	// value is what the SDK treats as "unset".
	b.Setenv(env.NexusBackendStore, "sqlite")
	b.Setenv(env.NexusDBSkipSchemaCreation, "")

	// Clean up the database
	dataDir := fs.NexusDataFolder()
	dbPath := filepath.Join(dataDir, "spike.db")
	if _, statErr := os.Stat(dbPath); statErr == nil {
		if rmErr := os.Remove(dbPath); rmErr != nil {
			b.Fatalf("Failed to remove existing database %s: %v", dbPath, rmErr)
		}
	}

	store := createTestDataStore(b)
	closeStoreOnCleanup(b, store)

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
