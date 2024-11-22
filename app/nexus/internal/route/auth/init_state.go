//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

func updateStateForInit(
	recoveryToken string, adminTokenBytes, salt []byte,
) error {
	iterationCount := env.Pbkdf2IterationCount()
	hashLength := env.ShaHashLength()
	recoveryTokenHash := pbkdf2.Key(
		[]byte(recoveryToken), salt,
		iterationCount, hashLength, sha256.New,
	)

	// As soon as SPIKE Nexus starts, it is guaranteed to have a root key in
	// memory. We don't need to fetch it from SPIKE Keeper. The only place
	// the root key is fetched from SPIKE Keeper is in the SPIKE Nexus init
	// flow.
	rootKey := state.RootKey()

	// Generate a 32-byte encryption key from the recovery token
	encryptionKey := pbkdf2.Key(
		[]byte(recoveryToken), salt,
		iterationCount, 32, sha256.New, // AES-256 requires 32-byte key
	)

	// Create new AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %v", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt root key
	encryptedRootKey := gcm.Seal(nonce, nonce, []byte(rootKey), nil)

	// TODO: we need a way to recover the metadata too :)
	state.SetAdminSigningToken("spike." + string(adminTokenBytes))
	state.SetAdminRecoveryMetadata(
		hex.EncodeToString(recoveryTokenHash),
		hex.EncodeToString(encryptedRootKey),
		hex.EncodeToString(salt),
	)

	return nil
}

// For completeness, here's the decryption function that would be used to recover the root key
func decryptRootKey(encryptedRootKeyHex, recoveryToken string, salt []byte) ([]byte, error) {
	// Decode hex-encoded encrypted key
	encryptedRootKey, err := hex.DecodeString(encryptedRootKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted root key: %v", err)
	}

	// Generate the same encryption key from recovery token
	iterationCount := env.Pbkdf2IterationCount()
	encryptionKey := pbkdf2.Key(
		[]byte(recoveryToken), salt,
		iterationCount, 32, sha256.New,
	)

	// Create cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Extract nonce from the encrypted data
	nonceSize := gcm.NonceSize()
	if len(encryptedRootKey) < nonceSize {
		return nil, fmt.Errorf("encrypted root key is too short")
	}

	nonce := encryptedRootKey[:nonceSize]
	ciphertext := encryptedRootKey[nonceSize:]

	// Decrypt
	rootKey, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root key: %v", err)
	}

	return rootKey, nil
}
