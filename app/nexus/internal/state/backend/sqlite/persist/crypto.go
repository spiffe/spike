//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/rand"
	"fmt"
	"io"
)

func (s *DataStore) encrypt(data []byte) ([]byte, []byte, error) {
	nonce := make([]byte, s.Cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := s.Cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nonce, nil
}

func (s *DataStore) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	plaintext, err := s.Cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return plaintext, nil
}
