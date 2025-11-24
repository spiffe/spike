//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/lite"
)

// initializeLiteBackend creates and initializes a Lite backend instance
// using the provided root key for encryption. The Lite backend is a
// lightweight alternative to SQLite for persistent storage. The Lite mode
// does not use any backing store and relies on persisting encrypted data
// on object storage (like S3, or Minio).
//
// Parameters:
//   - rootKey: A 32-byte encryption key used to secure the Lite database.
//     The backend will use this key directly for encryption operations.
//
// Returns:
//   - backend.Backend: An initialized Lite backend instance
//
// Error Handling:
// If the backend creation fails, the function calls log.FatalErr() which
// terminates the program. This is a fatal error because the system cannot
// operate without a properly initialized backend when Lite mode is
// configured.
//
// Example:
//
//	var rootKey [32]byte
//	// ... populate rootKey with secure random data ...
//	backend := initializeLiteBackend(&rootKey)
//	// backend is guaranteed to be valid; function exits on error
//
// Note: Unlike the SQLite backend, the Lite backend does not require a
// separate Initialize() call or timeout configuration.
func initializeLiteBackend(
	rootKey *[crypto.AES256KeySize]byte,
) backend.Backend {
	const fName = "initializeLiteBackend"

	dbBackend, err := lite.New(rootKey)
	if err != nil {
		failErr := sdkErrors.ErrStateInitializationFailed.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	return dbBackend
}
