//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"crypto/rand"
	stdErrors "errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/internal/net"
)

// decryptDataStreaming performs decryption for streaming mode requests.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data
//   - c: The cipher to use for decryption
//   - w: The HTTP response writer for error responses
//   - fName: The function name for logging
//
// Returns:
//   - plaintext: The decrypted data if successful
//   - error: An error if decryption fails
func decryptDataStreaming(
	nonce, ciphertext []byte, c cipher.AEAD, w http.ResponseWriter,
	fName string,
) ([]byte, error) {
	log.Log().Info(
		fName,
		"message", "decrypt",
		"len_nonce", len(nonce),
		"len_ciphertext", len(ciphertext),
	)

	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		http.Error(w, "decryption failed", http.StatusBadRequest)
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// decryptDataJSON performs decryption for JSON mode requests.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data
//   - c: The cipher to use for decryption
//   - w: The HTTP response writer for error responses
//   - fName: The function name for logging
//
// Returns:
//   - plaintext: The decrypted data if successful
//   - error: An error if decryption fails
func decryptDataJSON(
	nonce, ciphertext []byte, c cipher.AEAD, w http.ResponseWriter,
	fName string,
) ([]byte, error) {
	log.Log().Info(
		fName,
		"message", "decrypt",
		"len_nonce", len(nonce),
		"len_ciphertext", len(ciphertext),
	)

	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		failErr := stdErrors.Join(sdkErrors.ErrCryptoDecryptionFailed, err)
		return nil, net.Fail(
			reqres.CipherDecryptInternal, w, http.StatusInternalServerError,
			failErr, fName,
		)
	}

	return plaintext, nil
}

// generateNonceOrFailStreaming generates a cryptographically secure random
// nonce for streaming mode requests.
//
// Parameters:
//   - c: The cipher to determine nonce size
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - nonce: The generated nonce bytes if successful
//   - error: An error if nonce generation fails
func generateNonceOrFailStreaming(
	c cipher.AEAD, w http.ResponseWriter,
) ([]byte, error) {
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := stdErrors.Join(sdkErrors.ErrCryptoNonceGenerationFailed, err)
		http.Error(
			w, string(sdkErrors.ErrCodeCryptoNonceGenerationFailed),
			http.StatusInternalServerError,
		)
		return nil, failErr
	}

	return nonce, nil
}

// generateNonceOrFailJSON generates a cryptographically secure random nonce
// for JSON mode requests.
//
// Parameters:
//   - c: The cipher to determine nonce size
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//
// Returns:
//   - nonce: The generated nonce bytes if successful
//   - error: An error if nonce generation fails
func generateNonceOrFailJSON[T any](
	c cipher.AEAD, w http.ResponseWriter, errorResponse T,
) ([]byte, error) {
	const fName = "generateNonceOrFailJSON"

	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := stdErrors.Join(sdkErrors.ErrCryptoNonceGenerationFailed, err)
		return nil, net.Fail(
			errorResponse, w, http.StatusInternalServerError, failErr, fName,
		)
	}

	return nonce, nil
}

// encryptDataStreaming generates a nonce, performs encryption, and returns
// the nonce and ciphertext for streaming mode requests.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - c: The cipher to use for encryption
//   - w: The HTTP response writer for error responses
//   - fName: The function name for logging
//
// Returns:
//   - nonce: The generated nonce bytes
//   - ciphertext: The encrypted data
//   - error: An error if nonce generation fails
func encryptDataStreaming(
	plaintext []byte, c cipher.AEAD, w http.ResponseWriter, fName string,
) ([]byte, []byte, error) {
	nonce, err := generateNonceOrFailStreaming(c, w)
	if err != nil {
		return nil, nil, err
	}

	log.Log().Info(
		fName,
		"message", "encrypt",
		"len_nonce", len(nonce),
		"len_plaintext", len(plaintext),
	)
	ciphertext := c.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// encryptDataJSON generates a nonce, performs encryption, and returns the
// nonce and ciphertext for JSON mode requests.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - c: The cipher to use for encryption
//   - w: The HTTP response writer for error responses
//   - fName: The function name for logging
//
// Returns:
//   - nonce: The generated nonce bytes
//   - ciphertext: The encrypted data
//   - error: An error if nonce generation fails
func encryptDataJSON(
	plaintext []byte, c cipher.AEAD, w http.ResponseWriter, fName string,
) ([]byte, []byte, error) {
	nonce, err := generateNonceOrFailJSON(c, w, reqres.CipherEncryptInternal)
	if err != nil {
		return nil, nil, err
	}

	log.Log().Info(
		fName,
		"message", "encrypt",
		"len_nonce", len(nonce),
		"len_plaintext", len(plaintext),
	)
	ciphertext := c.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}
