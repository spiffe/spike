//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestingInterface allows both *testing.T and *testing.B to be used
type TestingInterface interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func createTestCipher(tb TestingInterface) cipher.AEAD {
	key := make([]byte, 32) // AES-256 key
	if _, err := rand.Read(key); err != nil {
		tb.Fatalf("Failed to generate test key: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		tb.Fatalf("Failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		tb.Fatalf("Failed to create GCM: %v", err)
	}

	return gcm
}

func createTestDataStore(tb TestingInterface, db *sql.DB) *DataStore {
	testCipher := createTestCipher(tb)
	return &DataStore{
		db:     db,
		Cipher: testCipher,
	}
}

func encryptTestData(tb TestingInterface, cipher cipher.AEAD, data map[string]string) ([]byte, []byte) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		tb.Fatalf("Failed to marshal test data: %v", err)
	}

	nonce := make([]byte, cipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		tb.Fatalf("Failed to generate nonce: %v", err)
	}

	encrypted := cipher.Seal(nil, nonce, jsonData, nil)
	return encrypted, nonce
}

func TestDataStore_loadSecretInternal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/secret/path"

	// Test data
	testData := map[string]string{
		"username": "admin",
		"password": "supersecret",
		"url":      "https://example.com",
	}

	encrypted, nonce := encryptTestData(t, store.Cipher, testData)

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)
	versionCreatedTime := time.Now().Add(-23 * time.Hour)

	// Mock metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Mock versions query
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
		AddRow(1, nonce, encrypted, versionCreatedTime, nil)
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

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
	if len(version.Data) != len(testData) {
		t.Errorf("Expected %d data items, got %d", len(testData), len(version.Data))
	}

	for key, expectedValue := range testData {
		actualValue, exists := version.Data[key]
		if !exists {
			t.Errorf("Expected key '%s' to exist", key)
		}
		if actualValue != expectedValue {
			t.Errorf("Expected '%s'='%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	if !version.CreatedTime.Equal(versionCreatedTime) {
		t.Errorf("Expected version created time %v, got %v", versionCreatedTime, version.CreatedTime)
	}

	if version.DeletedTime != nil {
		t.Error("Expected DeletedTime to be nil")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_MultipleVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/multi/versions"

	// Test data for different versions
	v1Data := map[string]string{"key": "value1", "env": "dev"}
	v2Data := map[string]string{"key": "value2", "env": "prod"}
	v3Data := map[string]string{"key": "value3", "env": "prod", "region": "us-east-1"}

	v1Encrypted, v1Nonce := encryptTestData(t, store.Cipher, v1Data)
	v2Encrypted, v2Nonce := encryptTestData(t, store.Cipher, v2Data)
	v3Encrypted, v3Nonce := encryptTestData(t, store.Cipher, v3Data)

	createdTime := time.Now().Add(-72 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)
	v1CreatedTime := time.Now().Add(-71 * time.Hour)
	v2CreatedTime := time.Now().Add(-25 * time.Hour)
	v3CreatedTime := time.Now().Add(-2 * time.Hour)
	v2DeletedTime := time.Now().Add(-1 * time.Hour)

	// Mock metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(3, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Mock versions query with multiple versions
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
		AddRow(1, v1Nonce, v1Encrypted, v1CreatedTime, nil).
		AddRow(2, v2Nonce, v2Encrypted, v2CreatedTime, v2DeletedTime).
		AddRow(3, v3Nonce, v3Encrypted, v3CreatedTime, nil)
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

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

	// Verify version 2 (deleted)
	version2 := secret.Versions[2]
	if version2.Data["key"] != "value2" {
		t.Errorf("Expected v2 key='value2', got '%s'", version2.Data["key"])
	}
	if version2.DeletedTime == nil {
		t.Error("Expected v2 DeletedTime to be set")
	} else if !version2.DeletedTime.Equal(v2DeletedTime) {
		t.Errorf("Expected v2 deleted time %v, got %v", v2DeletedTime, *version2.DeletedTime)
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

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_SecretNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "nonexistent/secret"

	// Mock metadata query to return no rows
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnError(sql.ErrNoRows)

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err != nil {
		t.Errorf("loadSecretInternal should not return error for non-existent secret: %v", err)
	}

	if secret != nil {
		t.Error("Expected nil secret for non-existent path")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_MetadataQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/error/path"

	// Mock metadata query to return database error
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnError(errors.New("database connection failed"))

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for database failure")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to load secret metadata"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_VersionsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/versions/error"

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)

	// Mock successful metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Mock versions query to return error
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnError(errors.New("versions table corrupted"))

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for versions query failure")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to query secret versions"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_VersionScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/scan/error"

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)

	// Mock successful metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Mock versions query with invalid data (wrong type for version)
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
		AddRow("invalid", []byte("nonce"), []byte("data"), time.Now(), nil)
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for scan failure")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to scan secret version"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_DecryptionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/decrypt/error"

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)
	versionCreatedTime := time.Now().Add(-23 * time.Hour)

	// Mock successful metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Create invalid encrypted data (wrong nonce/data combination)
	invalidNonce := make([]byte, store.Cipher.NonceSize())
	invalidEncrypted := []byte("invalid encrypted data")

	// Mock versions query with invalid encrypted data
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
		AddRow(1, invalidNonce, invalidEncrypted, versionCreatedTime, nil)
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for decryption failure")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to decrypt secret version"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_JSONUnmarshalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/json/error"

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)
	versionCreatedTime := time.Now().Add(-23 * time.Hour)

	// Mock successful metadata query
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Encrypt invalid JSON data
	invalidJSONData := []byte("invalid json data")
	nonce := make([]byte, store.Cipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}
	encrypted := store.Cipher.Seal(nil, nonce, invalidJSONData, nil)

	// Mock versions query with invalid JSON
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
		AddRow(1, nonce, encrypted, versionCreatedTime, nil)
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for JSON unmarshal failure")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to unmarshal secret values"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_EmptyVersionsResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx := context.Background()
	path := "test/empty/versions"

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now().Add(-1 * time.Hour)

	// Mock successful metadata query indicating version 1 exists
	metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
		AddRow(1, createdTime, updatedTime)
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnRows(metadataRows)

	// Mock versions query returning no rows (inconsistent state - metadata says version 1 exists but no versions found)
	versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"})
	mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
		WithArgs(path).
		WillReturnRows(versionRows)

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

	// Check that versions map is empty
	if len(secret.Versions) != 0 {
		t.Errorf("Expected no versions, got %d", len(secret.Versions))
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

func TestDataStore_loadSecretInternal_ContextTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	path := "test/timeout/context"

	// Mock metadata query to return context deadline exceeded error
	mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
		WithArgs(path).
		WillReturnError(context.DeadlineExceeded)

	// Execute the function
	secret, err := store.loadSecretInternal(ctx, path)

	// Verify results
	if err == nil {
		t.Error("Expected error for timeout context")
	}

	if secret != nil {
		t.Error("Expected nil secret on error")
	}

	expectedError := "failed to load secret metadata"
	if err != nil && !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Custom matcher for SQL arguments with time values
type timeArgMatcher struct {
	expected  time.Time
	tolerance time.Duration
}

func (m timeArgMatcher) Match(v driver.Value) bool {
	if t, ok := v.(time.Time); ok {
		diff := t.Sub(m.expected)
		if diff < 0 {
			diff = -diff
		}
		return diff <= m.tolerance
	}
	return false
}

// Benchmark test for performance
func BenchmarkDataStore_loadSecretInternal(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	store := createTestDataStore(b, db)
	ctx := context.Background()
	path := "benchmark/secret"

	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	encrypted, nonce := encryptTestData(b, store.Cipher, testData)
	createdTime := time.Now()
	updatedTime := time.Now()
	versionCreatedTime := time.Now()

	// Set up expectations that will be repeated
	for i := 0; i < b.N; i++ {
		metadataRows := sqlmock.NewRows([]string{"current_version", "created_time", "updated_time"}).
			AddRow(1, createdTime, updatedTime)
		mock.ExpectQuery("SELECT current_version, created_time, updated_time FROM secret_metadata WHERE path = ?").
			WithArgs(path).
			WillReturnRows(metadataRows)

		versionRows := sqlmock.NewRows([]string{"version", "nonce", "encrypted_data", "created_time", "deleted_time"}).
			AddRow(1, nonce, encrypted, versionCreatedTime, nil)
		mock.ExpectQuery("SELECT version, nonce, encrypted_data, created_time, deleted_time FROM secrets WHERE path = ?").
			WithArgs(path).
			WillReturnRows(versionRows)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.loadSecretInternal(ctx, path)
		if err != nil {
			b.Errorf("loadSecretInternal failed: %v", err)
		}
	}
}
