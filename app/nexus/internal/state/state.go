//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"sync"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/nexus/internal/net"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
	"github.com/spiffe/spike/internal/crypto"
)

var (
	rootKey   string
	rootKeyMu sync.RWMutex
)

var (
	adminToken   string
	adminTokenMu sync.RWMutex
)

var (
	kv   = store.NewKV()
	kvMu sync.RWMutex
)

// AdminToken returns the current admin token in a thread-safe manner.
// The returned token is protected by a read lock to ensure concurrent
// access safety.
func AdminToken() string {
	adminTokenMu.RLock()
	defer adminTokenMu.RUnlock()
	return adminToken
}

// SetAdminToken updates the admin token with the provided value.
// It uses a mutex to ensure thread-safe write operations.
//
// Parameters:
//   - token: The new admin token value to be set
func SetAdminToken(token string) {
	adminTokenMu.Lock()
	defer adminTokenMu.Unlock()
	adminToken = token
}

// UpsertSecret stores or updates a secret at the specified path with the
// provided values. It provides thread-safe access to the underlying key-value
// store.
//
// Parameters:
//   - path: The location where the secret should be stored
//   - values: A map containing the secret key-value pairs to be stored
func UpsertSecret(path string, values map[string]string) {
	kvMu.Lock()
	defer kvMu.Unlock()

	kv.Put(path, values)
}

// DeleteSecret deletes one or more versions of a secret at the specified path.
// It acquires a mutex lock before performing the deletion to ensure thread
// safety.
//
// Parameters:
//   - path: The path to the secret to be deleted
//   - versions: A slice of version numbers to delete. If empty, deletes the
//     current version only. Version number 0 is the current version.
func DeleteSecret(path string, versions []int) {
	kvMu.Lock()
	defer kvMu.Unlock()

	kv.Delete(path, versions)
}

// UndeleteSecret restores previously deleted versions of a secret at the
// specified path. It takes a path string identifying the secret's location and
// a slice of version numbers to restore. The function acquires a lock on the
// key-value store to ensure thread-safe operations during the undelete process.
//
// The function operates synchronously and will block until the undelete operation
// is complete. If any specified version numbers don't exist or were not previously
// deleted, those versions will be silently skipped.
//
// Parameters:
//   - path: The path to the secret to be restored
//   - versions: A slice of integer version numbers to restore
//
// Example:
//
//	// Restore versions 1 and 3 of a secret
//	UndeleteSecret("/app/secrets/api-key", []int{1, 3})
func UndeleteSecret(path string, versions []int) {
	kvMu.Lock()
	defer kvMu.Unlock()

	kv.Undelete(path, versions)
}

// ListKeys returns a slice of strings containing all keys currently stored
// in the key-value store. The function acquires a lock on the store to ensure
// a consistent view of the keys during enumeration.
//
// The returned slice contains the paths to all keys, regardless of their
// version status (active or deleted). The paths are returned in lexicographical
// order.
//
// Returns:
//   - []string: A slice containing all key paths in the store
//
// Example:
//
//	keys := ListKeys()
//	for _, key := range keys {
//	    fmt.Printf("Found key: %s\n", key)
//	}
func ListKeys() []string {
	kvMu.Lock()
	defer kvMu.Unlock()

	return kv.List()
}

// GetSecret retrieves a secret from the specified path and version.
// It provides thread-safe read access to the secret store.
//
// Parameters:
//   - path: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - map[string]string: The secret key-value pairs
//   - bool: Whether the secret was found
func GetSecret(path string, version int) (map[string]string, bool) {
	kvMu.RLock()
	defer kvMu.RUnlock()

	return kv.Get(path, version)
}

var ErrAlreadyInitialized = errors.New("already initialized")

// Initialize sets up the root key if it hasn't been initialized yet.
// If a root key already exists, it returns immediately.
// The root key is generated using AES-256 encryption.
//
// Returns:
//   - error: Any error encountered during initialization, nil on success
func Initialize(source *workloadapi.X509Source) error {
	existingRootKey := RootKey()
	if existingRootKey == "" {
		// Check if SPIKE Keeper has a cached root key first:
		key, err := net.FetchFromCache(source)
		if err != nil {
			return err
		}
		if key != "" {
			SetRootKey(key)
			existingRootKey = key
		}
	}

	if existingRootKey != "" { // if key empty, try getting it from SPIKE Keeper
		return ErrAlreadyInitialized
	}

	r, err := crypto.Aes256Seed()
	if err != nil {
		return err
	}

	rootKeyMu.Lock()
	rootKey = r
	rootKeyMu.Unlock()

	return nil
}

// RootKey returns the current root key in a thread-safe manner.
// The returned key is protected by a read lock to ensure concurrent
// access safety.
func RootKey() string {
	rootKeyMu.RLock()
	defer rootKeyMu.RUnlock()

	return rootKey
}

// SetRootKey sets the root key that is fetched from SPIKE Keeper.
func SetRootKey(key string) {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	rootKey = key
}
