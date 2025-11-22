//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/rand"
	"fmt"
	"io"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// generateNonce generates a cryptographically secure random nonce for use
// with AES-GCM encryption. The nonce size is determined by the cipher's
// requirements (typically 12 bytes for AES-GCM).
//
// If nonce generation fails, this function terminates the program via
// log.FatalErr, as this indicates a critical cryptographic system failure.
//
// Parameters:
//   - s: The DataStore containing the cipher whose nonce size will be used
//
// Returns:
//   - []byte: A cryptographically secure random nonce of the required size
//   - *sdkErrors.SDKError: Always returns nil (function terminates on error)
func generateNonce(s *DataStore) ([]byte, *sdkErrors.SDKError) {
	const fName = "generateNonce"

	nonce := make([]byte, s.Cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	return nonce, nil
}

// encryptWithNonce encrypts data using AES-GCM with the provided nonce.
// The function validates that the nonce size matches the cipher's
// requirements before performing encryption.
//
// Parameters:
//   - s: The DataStore containing the AES-GCM cipher for encryption
//   - nonce: The nonce to use for encryption (must match cipher's nonce size)
//   - data: The plaintext data to encrypt
//
// Returns:
//   - []byte: The encrypted ciphertext, or nil if an error occurs
//   - *sdkErrors.SDKError: nil on success, or ErrCryptoNonceGenerationFailed
//     if the nonce size does not match the cipher's requirements
func encryptWithNonce(
	s *DataStore, nonce []byte, data []byte,
) ([]byte, *sdkErrors.SDKError) {
	if len(nonce) != s.Cipher.NonceSize() {
		// TODO: this does not reflect the actual error; create an ErrCryptoNonceSizeMismatch instead.
		failErr := sdkErrors.ErrCryptoNonceGenerationFailed
		failErr.Msg = fmt.Sprintf(
			"invalid nonce size: got %d, want %d",
			len(nonce), s.Cipher.NonceSize(),
		)
		return nil, failErr
	}

	ciphertext := s.Cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nil
}
