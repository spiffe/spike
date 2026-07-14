//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/validation"
)

// InitializeBackend creates and returns a backend storage implementation based
// on the configured store type in the environment. The function is thread-safe
// through a mutex lock.
//
// Parameters:
//   - rootKey: The encryption key used for backend initialization (used by
//     SQLite backend)
//
// Returns:
//   - A backend.Backend interface implementation:
//   - memory.Store for 'memory' store type or unknown types
//   - SQLite backend for 'sqlite' store type
//
// The actual backend type is determined by env.BackendStoreType():
//   - env.Memory: Returns a no-op memory store
//   - env.Sqlite: Initializes and returns a SQLite backend
//   - default: Falls back to a no-op memory store
//
// The function is safe for concurrent access as it uses a mutex to protect the
// initialization process.
//
// Note: This function modifies the package-level be variable. Later calls
// will reinitialize the backend, potentially losing any existing state.
func InitializeBackend(rootKey *[crypto.AES256KeySize]byte) {
	// Root key is not needed, nor used in in-memory stores.
	// For in-memory stores, ensure that it is always nil, as the alternative
	// might mean a logic, or initialization-flow bug, and an unnecessary
	// crypto material in the memory.
	// In other store types, ensure it is set for security.
	if env.BackendStoreTypeVal() == env.Memory {
		validation.NilRootKeyOrDie(rootKey)
	} else {
		validation.ValidRootKeyOrDie(rootKey)
	}

	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreTypeVal()

	switch storeType {
	case env.Lite:
		be = initializeLiteBackend(rootKey)
	case env.Memory:
		be = initializeInMemoryBackend()
	case env.Sqlite:
		be = initializeSqliteBackend(rootKey)
	default:
		be = initializeInMemoryBackend()
	}

	// Store the backend atomically for safe concurrent access.
	backendPtr.Store(&be)
}
