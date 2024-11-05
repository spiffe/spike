//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import "sync"

var rootKey string
var rootKeyMutex sync.RWMutex

// RootKey returns the current root key value in a thread-safe manner.
// It uses a read lock to ensure concurrent read access is safe while
// preventing writes during the read operation.
func RootKey() string {
	rootKeyMutex.RLock()
	defer rootKeyMutex.RUnlock()
	return rootKey
}

// SetRootKey updates the root key value in a thread-safe manner.
// It acquires a write lock to ensure exclusive access during the update,
// preventing any concurrent reads or writes to the root key.
//
// Parameters:
//   - key: The new value to set as the root key
func SetRootKey(key string) {
	rootKeyMutex.Lock()
	defer rootKeyMutex.Unlock()
	rootKey = key
}
