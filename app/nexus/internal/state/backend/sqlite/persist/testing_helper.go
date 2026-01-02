//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// TestingInterface allows both *testing.T and *testing.B to be used
type TestingInterface interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

type TestSecretMetadata struct {
	CurrentVersion int
	OldestVersion  int
	MaxVersions    int
	CreatedTime    time.Time
	UpdatedTime    time.Time
}

// Helper functions for SQLite testing
func createTestRootKey(_ TestingInterface) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	// Use a predictable pattern for testing
	for i := range key {
		key[i] = byte(i % 256)
	}
	return key
}

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

	// Run the test function
	testFunc()
}

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

func createTestDataStore(t TestingInterface) *DataStore {
	rootKey := createTestRootKey(t)

	block, cipherErr := aes.NewCipher(rootKey[:])
	if cipherErr != nil {
		t.Fatalf("Failed to create cipher: %v", cipherErr)
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		t.Fatalf("Failed to create GCM: %v", gcmErr)
	}

	// Use DefaultOptions and override the data directory for testing
	opts := DefaultOptions()
	opts.DataDir = fs.NexusDataFolder()

	// Create a unique database filename to avoid race conditions
	opts.DatabaseFile = fmt.Sprintf("spike_test_%d.db", time.Now().UnixNano())

	store := &DataStore{
		Opts:   opts,
		Cipher: gcm,
	}

	// Initialize the database
	ctx := context.Background()
	if initErr := store.Initialize(ctx); initErr != nil {
		t.Fatalf("Failed to initialize datastore: %v", initErr)
	}

	dbPath := filepath.Join(opts.DataDir, opts.DatabaseFile)
	t.Logf("Test datastore initialized with database at %s", dbPath)
	return store
}

func storeTestSecretDirectly(t TestingInterface, store *DataStore, path string,
	versions map[int]map[string]string, metadata TestSecretMetadata) {
	ctx := context.Background()

	// Insert metadata
	_, metaErr := store.db.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
		path, metadata.CurrentVersion, metadata.OldestVersion,
		metadata.CreatedTime, metadata.UpdatedTime, metadata.MaxVersions)
	if metaErr != nil {
		t.Fatalf("Failed to insert metadata: %v", metaErr)
	}

	// Insert versions
	for version, data := range versions {
		// Encrypt the data
		jsonData := `{`
		first := true
		for k, v := range data {
			if !first {
				jsonData += `,`
			}
			jsonData += `"` + k + `":"` + v + `"`
			first = false
		}
		jsonData += `}`

		nonce := make([]byte, store.Cipher.NonceSize())
		if _, randErr := rand.Read(nonce); randErr != nil {
			t.Fatalf("Failed to generate nonce: %v", randErr)
		}

		encrypted := store.Cipher.Seal(nil, nonce, []byte(jsonData), nil)

		createdTime := metadata.CreatedTime.Add(time.Duration(version) * time.Hour)
		var deletedTime *time.Time
		if version == 2 {
			// Make version 2 deleted for testing
			deleted := metadata.UpdatedTime.Add(-1 * time.Hour)
			deletedTime = &deleted
		}

		_, execErr := store.db.ExecContext(ctx, ddl.QueryUpsertSecret,
			path, version, nonce, encrypted, createdTime, deletedTime)
		if execErr != nil {
			t.Fatalf("Failed to insert version %d: %v", version, execErr)
		}
	}
}
