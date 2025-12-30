//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike-sdk-go/net"
)

// decryptDataStreaming performs decryption for streaming mode requests.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data
//   - c: The cipher to use for decryption
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - plaintext: The decrypted data if successful
//   - *sdkErrors.SDKError: An error if decryption fails
func decryptDataStreaming(
	nonce, ciphertext []byte, c cipher.AEAD, w http.ResponseWriter,
) ([]byte, *sdkErrors.SDKError) {
	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		http.Error(w, "decryption failed", http.StatusBadRequest)
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
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
//
// Returns:
//   - plaintext: The decrypted data if successful
//   - *sdkErrors.SDKError: An error if decryption fails
func decryptDataJSON(
	nonce, ciphertext []byte, c cipher.AEAD, w http.ResponseWriter,
) ([]byte, *sdkErrors.SDKError) {
	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		failErr := net.Fail(
			reqres.CipherDecryptResponse{}.Internal(), w,
			http.StatusInternalServerError,
		)
		if failErr != nil {
			return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(err).Wrap(failErr)
		}
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
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
//   - *sdkErrors.SDKError: An error if nonce generation fails
func generateNonceOrFailStreaming(
	c cipher.AEAD, w http.ResponseWriter,
) ([]byte, *sdkErrors.SDKError) {
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		http.Error(
			w, string(sdkErrors.ErrCryptoNonceGenerationFailed.Code),
			http.StatusInternalServerError,
		)
		return nil, sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(err)
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
//   - *sdkErrors.SDKError: An error if nonce generation fails
func generateNonceOrFailJSON[T any](
	c cipher.AEAD, w http.ResponseWriter, errorResponse T,
) ([]byte, *sdkErrors.SDKError) {
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := net.Fail(errorResponse, w, http.StatusInternalServerError)
		if failErr != nil {
			return nil, sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(
				err).Wrap(failErr)
		}
		return nil, sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(err)
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
//
// Returns:
//   - nonce: The generated nonce bytes
//   - ciphertext: The encrypted data
//   - *sdkErrors.SDKError: An error if nonce generation fails
func encryptDataStreaming(
	plaintext []byte, c cipher.AEAD, w http.ResponseWriter,
) ([]byte, []byte, *sdkErrors.SDKError) {
	nonce, err := generateNonceOrFailStreaming(c, w)
	if err != nil {
		return nil, nil, err
	}

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
//
// Returns:
//   - nonce: The generated nonce bytes
//   - ciphertext: The encrypted data
//   - *sdkErrors.SDKError: An error if nonce generation fails
func encryptDataJSON(
	plaintext []byte, c cipher.AEAD, w http.ResponseWriter,
) ([]byte, []byte, *sdkErrors.SDKError) {
	nonce, err := generateNonceOrFailJSON(
		c, w, reqres.CipherEncryptResponse{}.Internal(),
	)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := c.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}
