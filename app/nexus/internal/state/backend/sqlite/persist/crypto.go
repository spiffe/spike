//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/rand"
	"io"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// TODO: these are generic enough to move to the SDK.

// encrypt encrypts the given data using the DataStore's cipher.
// It generates a random nonce for each encryption operation to ensure
// uniqueness.
//
// Parameters:
//   - data: The plaintext data to encrypt
//
// Returns:
//   - []byte: The encrypted ciphertext
//   - []byte: The generated nonce used for encryption
//   - *sdkErrors.SDKError: nil on success, or
//     sdkErrors.ErrCryptoNonceGenerationFailed if nonce generation fails
func (s *DataStore) encrypt(data []byte) ([]byte, []byte, *sdkErrors.SDKError) {
	nonce := make([]byte, s.Cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(err)
		return nil, nil, failErr
	}
	ciphertext := s.Cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nonce, nil
}

// decrypt decrypts the given ciphertext using the DataStore's cipher
// and the provided nonce.
//
// Parameters:
//   - ciphertext: The encrypted data to decrypt
//   - nonce: The nonce that was used during encryption
//
// Returns:
//   - []byte: The decrypted plaintext data
//   - *sdkErrors.SDKError: nil on success, or
//     sdkErrors.ErrCryptoDecryptionFailed if decryption fails
func (s *DataStore) decrypt(
	ciphertext, nonce []byte,
) ([]byte, *sdkErrors.SDKError) {
	plaintext, err := s.Cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
		return nil, failErr
	}
	return plaintext, nil
}
