//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
)

const (
	// AEADAlgAES256GCM is the AEAD algorithm identifier
	AEADAlgAES256GCM = "AES-256-GCM"
)

// WrapResult contains the result of wrapping a DEK with a KEK
type WrapResult struct {
	// KekID is the ID of the KEK used for wrapping
	KekID string

	// WrappedDEK is the encrypted DEK
	WrappedDEK []byte

	// Nonce used for encryption
	Nonce []byte

	// Tag for authentication (if separate from ciphertext)
	Tag []byte

	// AAD is additional authenticated data
	AAD []byte

	// AEADAlg is the algorithm used
	AEADAlg string
}

// UnwrapResult contains the result of unwrapping a DEK
type UnwrapResult struct {
	// DEK is the unwrapped Data Encryption Key
	DEK *[crypto.AES256KeySize]byte

	// KekID is the KEK that was used
	KekID string
}

// WrapDEK wraps a DEK with the specified KEK using AES-GCM
//
// Parameters:
//   - dek: The Data Encryption Key to wrap
//   - kek: The Key Encryption Key to use for wrapping
//   - kekID: The ID of the KEK
//   - aad: Optional additional authenticated data
//
// Returns:
//   - WrapResult containing the wrapped DEK and metadata
//   - Error if wrapping fails
func WrapDEK(
	dek *[crypto.AES256KeySize]byte,
	kek *[crypto.AES256KeySize]byte,
	kekID string,
	aad []byte,
) (*WrapResult, error) {
	const fName = "WrapDEK"

	if dek == nil || kek == nil {
		return nil, fmt.Errorf("%s: nil DEK or KEK", fName)
	}

	// Create AES cipher
	block, err := aes.NewCipher(kek[:])
	if err != nil {
		log.Log().Error(fName, "message", "failed to create cipher", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to create cipher: %w", fName, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Log().Error(fName, "message", "failed to create GCM", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to create GCM: %w", fName, err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Log().Error(fName, "message", "failed to generate nonce", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to generate nonce: %w", fName, err)
	}

	// Encrypt DEK
	// The Seal method appends the ciphertext and tag together
	wrappedDEK := gcm.Seal(nil, nonce, dek[:], aad)

	return &WrapResult{
		KekID:      kekID,
		WrappedDEK: wrappedDEK,
		Nonce:      nonce,
		AAD:        aad,
		AEADAlg:    AEADAlgAES256GCM,
	}, nil
}

// UnwrapDEK unwraps a DEK that was wrapped with a KEK
//
// Parameters:
//   - wrappedDEK: The encrypted DEK
//   - kek: The Key Encryption Key to use for unwrapping
//   - kekID: The ID of the KEK
//   - nonce: The nonce used during encryption
//   - aad: Additional authenticated data (must match what was used during wrap)
//
// Returns:
//   - UnwrapResult containing the DEK
//   - Error if unwrapping fails
func UnwrapDEK(
	wrappedDEK []byte,
	kek *[crypto.AES256KeySize]byte,
	kekID string,
	nonce []byte,
	aad []byte,
) (*UnwrapResult, error) {
	const fName = "UnwrapDEK"

	if kek == nil {
		return nil, fmt.Errorf("%s: nil KEK", fName)
	}

	// Create AES cipher
	block, err := aes.NewCipher(kek[:])
	if err != nil {
		log.Log().Error(fName, "message", "failed to create cipher", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to create cipher: %w", fName, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Log().Error(fName, "message", "failed to create GCM", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to create GCM: %w", fName, err)
	}

	// Decrypt DEK
	dekBytes, err := gcm.Open(nil, nonce, wrappedDEK, aad)
	if err != nil {
		log.Log().Error(fName, "message", "failed to unwrap DEK", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to unwrap DEK: %w", fName, err)
	}

	// Validate DEK size
	if len(dekBytes) != crypto.AES256KeySize {
		return nil, fmt.Errorf("%s: invalid DEK size: %d", fName, len(dekBytes))
	}

	// Copy to fixed-size array
	dek := new([crypto.AES256KeySize]byte)
	copy(dek[:], dekBytes)

	return &UnwrapResult{
		DEK:   dek,
		KekID: kekID,
	}, nil
}

// RewrapDEK unwraps a DEK with an old KEK and rewraps it with a new KEK
//
// This is used during lazy rewrap operations to upgrade secrets to use the
// current KEK without re-encrypting the actual secret data.
//
// Parameters:
//   - oldWrappedDEK: The DEK wrapped with the old KEK
//   - oldKek: The old Key Encryption Key
//   - oldKekID: The ID of the old KEK
//   - oldNonce: The nonce used with the old KEK
//   - oldAAD: AAD used with the old KEK
//   - newKek: The new Key Encryption Key
//   - newKekID: The ID of the new KEK
//   - newAAD: AAD to use with the new KEK (can be same as old)
//
// Returns:
//   - WrapResult containing the rewrapped DEK
//   - Error if rewrapping fails
func RewrapDEK(
	oldWrappedDEK []byte,
	oldKek *[crypto.AES256KeySize]byte,
	oldKekID string,
	oldNonce []byte,
	oldAAD []byte,
	newKek *[crypto.AES256KeySize]byte,
	newKekID string,
	newAAD []byte,
) (*WrapResult, error) {
	const fName = "RewrapDEK"

	// Unwrap with old KEK
	unwrapResult, err := UnwrapDEK(oldWrappedDEK, oldKek, oldKekID, oldNonce, oldAAD)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to unwrap: %w", fName, err)
	}

	// Wrap with new KEK
	wrapResult, err := WrapDEK(unwrapResult.DEK, newKek, newKekID, newAAD)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to wrap: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "successfully rewrapped DEK",
		"old_kek_id", oldKekID,
		"new_kek_id", newKekID)

	return wrapResult, nil
}

// EncryptWithDEK encrypts data using a DEK
//
// Parameters:
//   - plaintext: The data to encrypt
//   - dek: The Data Encryption Key
//
// Returns:
//   - ciphertext: The encrypted data
//   - nonce: The nonce used for encryption
//   - Error if encryption fails
func EncryptWithDEK(
	plaintext []byte,
	dek *[crypto.AES256KeySize]byte,
) (ciphertext []byte, nonce []byte, err error) {
	const fName = "EncryptWithDEK"

	if dek == nil {
		return nil, nil, fmt.Errorf("%s: nil DEK", fName)
	}

	// Create AES cipher
	block, err := aes.NewCipher(dek[:])
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create cipher: %w", fName, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create GCM: %w", fName, err)
	}

	// Generate nonce
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("%s: failed to generate nonce: %w", fName, err)
	}

	// Encrypt
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// DecryptWithDEK decrypts data using a DEK
//
// Parameters:
//   - ciphertext: The encrypted data
//   - dek: The Data Encryption Key
//   - nonce: The nonce used during encryption
//
// Returns:
//   - plaintext: The decrypted data
//   - Error if decryption fails
func DecryptWithDEK(
	ciphertext []byte,
	dek *[crypto.AES256KeySize]byte,
	nonce []byte,
) ([]byte, error) {
	const fName = "DecryptWithDEK"

	if dek == nil {
		return nil, fmt.Errorf("%s: nil DEK", fName)
	}

	// Create AES cipher
	block, err := aes.NewCipher(dek[:])
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create cipher: %w", fName, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create GCM: %w", fName, err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to decrypt: %w", fName, err)
	}

	return plaintext, nil
}

// BuildSecretMetadata constructs SecretMetadata from a WrapResult
func BuildSecretMetadata(wrap *WrapResult) *SecretMetadata {
	now := time.Now()
	return &SecretMetadata{
		KekID:       wrap.KekID,
		WrappedDEK:  wrap.WrappedDEK,
		AEADAlg:     wrap.AEADAlg,
		Nonce:       wrap.Nonce,
		AAD:         wrap.AAD,
		RewrappedAt: &now,
	}
}
