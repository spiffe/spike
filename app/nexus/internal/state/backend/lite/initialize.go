//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package lite

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/noop"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

// Store implements the backend.Backend interface, providing encryption
// without persistent storage.
//
// This store embeds noop.Store for all storage operations (which are no-ops)
// and provides AES-GCM encryption capabilities through its Cipher field. It
// acts as an encryption-as-a-service layer, suitable for scenarios where
// encryption is required but data persistence is handled in-memory or by
// another component.
type Store struct {
	noop.Store             // Embedded no-op store for storage operations
	Cipher     cipher.AEAD // AES-GCM cipher for data encryption/decryption
}

// New creates a new lite backend with AES-GCM encryption.
//
// This function initializes an AES cipher block using the provided root key
// and wraps it with GCM (Galois/Counter Mode) for authenticated encryption.
// The resulting backend provides encryption services without any persistent
// storage functionality.
//
// Parameters:
//   - rootKey: A 256-bit (32-byte) AES key used for encryption/decryption
//
// Returns:
//   - backend.Backend: An initialized lite backend with AES-GCM encryption
//   - *sdkErrors.SDKError: nil on success, or one of the following errors:
//   - sdkErrors.ErrCryptoFailedToCreateCipher if AES cipher creation fails
//   - sdkErrors.ErrCryptoFailedToCreateGCM if GCM mode initialization fails
func New(rootKey *[crypto.AES256KeySize]byte) (
	backend.Backend, *sdkErrors.SDKError,
) {
	block, cipherErr := aes.NewCipher(rootKey[:])
	if cipherErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateCipher.Wrap(cipherErr)
		return nil, failErr
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateGCM.Wrap(gcmErr)
		return nil, failErr
	}

	return &Store{
		Cipher: gcm,
	}, nil
}

// GetCipher returns the AES-GCM cipher used for data encryption and
// decryption.
//
// This method provides access to the underlying AEAD (Authenticated Encryption
// with Associated Data) cipher for performing cryptographic operations.
//
// This method includes a nil receiver guard because it may be passed as a
// method value (e.g., `backend.GetCipher` without parentheses) where the
// receiver is bound at capture time. If the backend is nil when the method
// value is later invoked, the guard prevents a panic by returning nil
// instead of dereferencing a nil pointer. Callers should check for a nil
// return value.
//
// Returns:
//   - cipher.AEAD: The AES-GCM cipher instance configured during store
//     initialization, or nil if the receiver is nil
func (ds *Store) GetCipher() cipher.AEAD {
	if ds == nil {
		return nil
	}
	return ds.Cipher
}
