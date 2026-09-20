//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

func createTestCipher(t *testing.T) cipher.AEAD {
	key := make([]byte, 32) // AES-256 key
	if _, randErr := rand.Read(key); randErr != nil {
		t.Fatalf("Failed to generate test key: %v", randErr)
	}

	block, cipherErr := aes.NewCipher(key)
	if cipherErr != nil {
		t.Fatalf("Failed to create cipher: %v", cipherErr)
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		t.Fatalf("Failed to create GCM: %v", gcmErr)
	}

	return gcm
}
