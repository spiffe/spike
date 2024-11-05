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
