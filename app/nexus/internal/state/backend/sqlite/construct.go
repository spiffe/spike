//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/persist"
)

// New creates a new DataStore instance with the provided configuration.
// It validates the encryption key and initializes the AES-GCM cipher.
//
// The encryption key must be exactly 32 bytes in length (AES-256).
//
// Parameters:
//   - cfg: The backend configuration containing encryption key and options
//
// Returns:
//   - backend.Backend: The initialized SQLite backend on success
//   - *sdkErrors.SDKError: An error if initialization fails
//
// Errors returned:
//   - ErrStoreInvalidConfiguration: If options are invalid or key is
//     malformed
//   - ErrCryptoInvalidEncryptionKeyLength: If key is not 32 bytes
//   - ErrCryptoFailedToCreateCipher: If AES cipher creation fails
//   - ErrCryptoFailedToCreateGCM: If GCM mode initialization fails
func New(cfg backend.Config) (backend.Backend, *sdkErrors.SDKError) {
	opts, err := persist.ParseOptions(cfg.Options)
	if err != nil {
		failErr := sdkErrors.ErrStoreInvalidConfiguration.Wrap(err)
		return nil, failErr
	}

	key, decodeErr := hex.DecodeString(cfg.EncryptionKey)
	if decodeErr != nil {
		failErr := sdkErrors.ErrStoreInvalidConfiguration.Wrap(decodeErr)
		failErr.Msg = "invalid encryption key"
		return nil, failErr
	}

	// Validate key length
	if len(key) != crypto.AES256KeySize {
		failErr := *sdkErrors.ErrCryptoInvalidEncryptionKeyLength // copy
		failErr.Msg = "encryption key must be exactly 32 bytes"
		return nil, &failErr
	}

	block, aesErr := aes.NewCipher(key)
	if aesErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateCipher.Wrap(aesErr)
		failErr.Msg = "failed to create AES cipher"
		return nil, failErr
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateGCM.Wrap(gcmErr)
		failErr.Msg = "failed to create GCM mode"
		return nil, failErr
	}

	return &persist.DataStore{
		Cipher: gcm,
		Opts:   opts,
	}, nil
}
