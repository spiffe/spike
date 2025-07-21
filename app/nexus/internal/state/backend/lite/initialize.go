//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package lite

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
)

// DataStore implements the backend.Backend interface providing encryption
// without storage. It uses AES-GCM
type DataStore struct {
	memory.NoopStore
	Cipher cipher.AEAD // Encryption Cipher for data protection
}

// New creates a new Backend with AES-GCM encryption using the provided key.
// Returns an error if cipher initialization fails.
func New(rootKey *[32]byte) (backend.Backend, error) {
	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &DataStore{
		Cipher: gcm,
	}, nil
}

// GetCipher returns the encryption cipher used for data protection.
func (ds *DataStore) GetCipher() cipher.AEAD {
	return ds.Cipher
}
