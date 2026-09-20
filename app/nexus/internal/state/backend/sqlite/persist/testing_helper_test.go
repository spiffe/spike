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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

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

// withSQLiteEnvironment runs a function with the SQLite backend selected and
// schema creation enabled.
//
// t.Setenv restores the previous values when the test ends. An empty value
// is what the SDK treats as "unset" for both variables.
//
// Parameters:
//   - t: The test whose environment is modified.
//   - testFunc: The function to run with the environment in place.
func withSQLiteEnvironment(t *testing.T, testFunc func()) {
	t.Helper()
	t.Setenv(env.NexusBackendStore, "sqlite")
	t.Setenv(env.NexusDBSkipSchemaCreation, "")

	testFunc()
}

// closeStoreOnCleanup closes the store when the test or benchmark ends and
// fails it if Close reports an error.
//
// Parameters:
//   - tb: The test or benchmark that owns the store.
//   - store: The data store to close during cleanup.
func closeStoreOnCleanup(tb testing.TB, store *DataStore) {
	tb.Helper()
	tb.Cleanup(func() {
		if closeErr := store.Close(context.Background()); closeErr != nil {
			tb.Errorf("failed to close the datastore: %v", closeErr)
		}
	})
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

	insertEncryptedMetadata(ctx, t, store, path, metadata)

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

		_, execErr := store.db.ExecContext(ctx, ddl.QueryUpsertVersion,
			path, version, nonce, encrypted, createdTime, deletedTime)
		if execErr != nil {
			t.Fatalf("Failed to insert version %d: %v", version, execErr)
		}
	}
}

// insertEncryptedMetadata encrypts typed metadata fields with per-field
// nonces and inserts the encrypted values into the secret_metadata table.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - t: The test used to report failures.
//   - store: The data store whose database and cipher are used.
//   - path: The secret path the metadata row belongs to.
//   - metadata: The typed metadata values to encrypt and store.
func insertEncryptedMetadata(
	ctx context.Context, t TestingInterface, store *DataStore,
	path string, metadata TestSecretMetadata,
) {
	insertRawEncryptedMetadata(ctx, t, store, path, rawSecretMetadata{
		currentVersion: strconv.Itoa(metadata.CurrentVersion),
		oldestVersion:  strconv.Itoa(metadata.OldestVersion),
		maxVersions:    strconv.Itoa(metadata.MaxVersions),
		createdTime:    strconv.FormatInt(metadata.CreatedTime.Unix(), 10),
		updatedTime:    strconv.FormatInt(metadata.UpdatedTime.Unix(), 10),
	})
}

// rawSecretMetadata holds the plaintext strings that get encrypted into a
// secret_metadata row.
//
// Tests use it to write values that a healthy SPIKE would never produce,
// such as a non-numeric version, so that the loader's integrity checks can
// be exercised.
type rawSecretMetadata struct {
	currentVersion string
	oldestVersion  string
	maxVersions    string
	createdTime    string
	updatedTime    string
}

// insertRawEncryptedMetadata encrypts plaintext metadata fields with
// per-field derived nonces and upserts them into the secret_metadata table.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - t: The test used to report failures.
//   - store: The data store whose database and cipher are used.
//   - path: The secret path the metadata row belongs to.
//   - raw: The plaintext field values to encrypt and store.
func insertRawEncryptedMetadata(
	ctx context.Context, t TestingInterface, store *DataStore,
	path string, raw rawSecretMetadata,
) {
	nonce := make([]byte, store.Cipher.NonceSize())
	if _, randErr := rand.Read(nonce); randErr != nil {
		t.Fatalf("Failed to generate metadata nonce: %v", randErr)
	}

	encrypt := func(field, plaintext string) []byte {
		encrypted, encryptErr := encryptWithDerivedNonce(
			store, nonce, field, []byte(plaintext),
		)
		if encryptErr != nil {
			t.Fatalf("Failed to encrypt %s: %v", field, encryptErr)
		}
		return encrypted
	}

	encryptedCurrentVersion := encrypt(
		nonceFieldSecretMetadataCurrentVersion, raw.currentVersion)
	encryptedOldestVersion := encrypt(
		nonceFieldSecretMetadataOldestVersion, raw.oldestVersion)
	encryptedMaxVersions := encrypt(
		nonceFieldSecretMetadataMaxVersions, raw.maxVersions)
	encryptedCreatedTime := encrypt(
		nonceFieldSecretMetadataCreatedTime, raw.createdTime)
	encryptedUpdatedTime := encrypt(
		nonceFieldSecretMetadataUpdatedTime, raw.updatedTime)

	_, execErr := store.db.ExecContext(ctx, ddl.QueryUpsertMetadata,
		path, nonce, encryptedCurrentVersion, encryptedOldestVersion,
		encryptedCreatedTime, encryptedUpdatedTime, encryptedMaxVersions,
	)
	if execErr != nil {
		t.Fatalf("Failed to insert metadata: %v", execErr)
	}
}

// newMemoryDataStore opens an in-memory SQLite database wrapped in a
// DataStore, closed automatically when the test ends.
//
// Parameters:
//   - t: The test that owns the store; failed if the database cannot open.
//
// Returns:
//   - *DataStore: A store backed by an in-memory database.
func newMemoryDataStore(t *testing.T) *DataStore {
	db, openErr := sql.Open("sqlite3", ":memory:")
	if openErr != nil {
		t.Fatalf("failed to open an in-memory database: %v", openErr)
		return nil
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close the in-memory database: %v", closeErr)
		}
	})
	return &DataStore{db: db}
}

// busyErr mimics how a SQLITE_BUSY failure ("database is locked" in the
// mattn/go-sqlite3 driver) surfaces from the persist layer: wrapped in
// an SDKError chain.
//
// Returns:
//   - *sdkErrors.SDKError: A query failure wrapping the driver's busy error.
func busyErr() *sdkErrors.SDKError {
	return sdkErrors.ErrEntityQueryFailed.Wrap(
		errors.New("database is locked"),
	)
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
