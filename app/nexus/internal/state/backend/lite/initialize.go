//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lite provides an encryption-only backend that does not persist data.
//
// This package implements a lightweight backend that provides AES-GCM
// encryption services without any storage functionality. It embeds the noop
// backend for storage operations and only provides encryption capabilities.
// This is useful when encryption is needed, but data persistence is handled
// in-memory or by another layer.
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
	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateCipher.Wrap(err)
		return nil, failErr
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateGCM.Wrap(err)
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
// Returns:
//   - cipher.AEAD: The AES-GCM cipher instance configured during store
//     initialization
func (ds *Store) GetCipher() cipher.AEAD {
	return ds.Cipher
}
