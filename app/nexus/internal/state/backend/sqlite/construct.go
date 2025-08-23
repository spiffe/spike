//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/persist"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

// New creates a new DataStore instance with the provided configuration.
// It validates the encryption key and initializes the AES-GCM cipher.
//
// The encryption key must be 16, 24, or 32 bytes in length (for AES-128,
// AES-192, or AES-256 respectively).
//
// Returns an error if:
// - The options are invalid
// - The encryption key is malformed or has an invalid length
// - The cipher initialization fails
func New(cfg backend.Config) (backend.Backend, error) {
	opts, err := persist.ParseOptions(cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("invalid sqlite options: %w", err)
	}

	key, err := hex.DecodeString(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	// Validate key length
	if len(key) != crypto.AES256KeySize {
		return nil, fmt.Errorf(
			"invalid encryption key length: must be 32 bytes",
		)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &persist.DataStore{
		Cipher: gcm,
		Opts:   opts,
	}, nil
}
