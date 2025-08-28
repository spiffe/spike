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

	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
	"github.com/spiffe/spike/internal/config"
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
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	originalSkipSchema := os.Getenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	// Ensure cleanup happens
	defer func() {
		if originalStore != "" {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
		if originalSkipSchema != "" {
			_ = os.Setenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION", originalSkipSchema)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")
		}
	}()

	// Set to SQLite backend and ensure schema creation
	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "sqlite")
	_ = os.Unsetenv("SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION")

	// Run the test function
	testFunc()
}

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

func createTestDataStore(t TestingInterface) *DataStore {
	rootKey := createTestRootKey(t)

	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		t.Fatalf("Failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("Failed to create GCM: %v", err)
	}

	// Use DefaultOptions and override the data directory for testing
	opts := DefaultOptions()
	opts.DataDir = config.SpikeNexusDataFolder()

	// Create a unique database filename to avoid race conditions
	opts.DatabaseFile = fmt.Sprintf("spike_test_%d.db", time.Now().UnixNano())

	store := &DataStore{
		Opts:   opts,
		Cipher: gcm,
	}

	// Initialize the database
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize datastore: %v", err)
	}

	dbPath := filepath.Join(opts.DataDir, opts.DatabaseFile)
	t.Logf("Test datastore initialized with database at %s", dbPath)
	return store
}

func storeTestSecretDirectly(t TestingInterface, store *DataStore, path string, versions map[int]map[string]string, metadata TestSecretMetadata) {
	ctx := context.Background()

	// Insert metadata
	_, err := store.db.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
		path, metadata.CurrentVersion, metadata.OldestVersion,
		metadata.CreatedTime, metadata.UpdatedTime, metadata.MaxVersions)
	if err != nil {
		t.Fatalf("Failed to insert metadata: %v", err)
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
		if _, err := rand.Read(nonce); err != nil {
			t.Fatalf("Failed to generate nonce: %v", err)
		}

		encrypted := store.Cipher.Seal(nil, nonce, []byte(jsonData), nil)

		createdTime := metadata.CreatedTime.Add(time.Duration(version) * time.Hour)
		var deletedTime *time.Time
		if version == 2 {
			// Make version 2 deleted for testing
			deleted := metadata.UpdatedTime.Add(-1 * time.Hour)
			deletedTime = &deleted
		}

		_, err := store.db.ExecContext(ctx, ddl.QueryUpsertSecret,
			path, version, nonce, encrypted, createdTime, deletedTime)
		if err != nil {
			t.Fatalf("Failed to insert version %d: %v", version, err)
		}
	}
}
