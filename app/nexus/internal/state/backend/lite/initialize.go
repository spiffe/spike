//    \\ SPIKE: Secure your secrets with SPIFFE.
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

func New(rootKey *[32]byte) (backend.Backend, error) {
	if len(rootKey) != 32 {
		return nil, fmt.Errorf(
			"invalid encryption key length: must be 32 bytes",
		)
	}

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

func (ds *DataStore) GetCipher() cipher.AEAD {
	return ds.Cipher
}
