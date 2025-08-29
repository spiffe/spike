//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
)

var be backend.Backend

func createCipher() cipher.AEAD {
	key := make([]byte, crypto.AES256KeySize) // AES-256 key
	if _, err := rand.Read(key); err != nil {
		log.FatalLn("createCipher", "message", "Failed to generate test key", "err", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.FatalLn("createCipher", "message", "Failed to create cipher", "err", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.FatalLn("createCipher", "message", "Failed to create GCM", "err", err)
	}

	return gcm
}

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
// Note: This function modifies the package-level be variable. Subsequent calls
// will reinitialize the backend, potentially losing any existing state.
func InitializeBackend(rootKey *[crypto.AES256KeySize]byte) {
	const fName = "initializeBackend"

	log.Log().Info(fName,
		"message", "Initializing backend", "storeType", env.BackendStoreType())

	// Root key is not needed, nor used in in-memory stores.
	// For in-memory stores, ensure that it is always nil, as the alternative
	// might mean a logic, or initialization-flow bug, and an unnecessary
	// crypto material in the memory.
	// In other store types, ensure it is set for security.
	if env.BackendStoreType() == env.Memory {
		if rootKey != nil {
			log.FatalLn(fName,
				"message", "In-memory store can only be initialized with nil root key",
				"err", "root key is not nil",
			)
		}
	} else {
		if rootKey == nil {
			log.FatalLn(fName,
				"message", "Failed to initialize backend",
				"err", "root key is nil",
			)
		}

		if mem.Zeroed32(rootKey) {
			log.FatalLn(fName,
				"message", "Failed to initialize backend",
				"err", "root key is all zeroes",
			)
		}
	}

	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Lite:
		be = initializeLiteBackend(rootKey)
	case env.Memory:
		be = memory.NewInMemoryStore(createCipher(), env.MaxSecretVersions())
	case env.Sqlite:
		be = initializeSqliteBackend(rootKey)
	default:
		be = memory.NewInMemoryStore(createCipher(), env.MaxSecretVersions())
	}

	log.Log().Info(
		fName, "message", "Backend initialized", "storeType", storeType,
	)
}
