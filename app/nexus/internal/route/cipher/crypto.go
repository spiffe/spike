//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/internal/net"
)

// decryptData performs the actual decryption operation and handles errors
// appropriately based on the request mode.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data
//   - c: The cipher to use for decryption
//   - w: The HTTP response writer for error responses
//   - streamModeActive: Whether the request is in streaming mode
//   - fName: The function name for logging
//
// Returns:
//   - plaintext: The decrypted data if successful
//   - error: An error if decryption fails
func decryptData(
	nonce, ciphertext []byte, c cipher.AEAD, w http.ResponseWriter,
	streamModeActive bool, fName string,
) ([]byte, error) {
	log.Log().Info(fName, "message",
		fmt.Sprintf("Decrypt %d %d", len(nonce), len(ciphertext)),
	)

	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		if streamModeActive {
			http.Error(w, "decryption failed", http.StatusBadRequest)
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		return nil, net.Fail(
			reqres.CipherDecryptResponse{Err: data.ErrInternal},
			w,
			http.StatusInternalServerError,
			fmt.Errorf("decryption failed: %w", err),
			fName,
		)
	}

	return plaintext, nil
}

// generateNonceOrFail generates a cryptographically secure random nonce and
// handles errors appropriately based on the request mode.
//
// If nonce generation fails, the function sends an appropriate error response
// to the client based on whether streaming mode is active:
//   - Streaming mode: Sends plain HTTP error
//   - JSON mode: Sends structured JSON error response
//
// Parameters:
//   - c: The cipher to determine nonce size
//   - w: The HTTP response writer for error responses
//   - streamModeActive: Whether the request is in streaming mode
//   - errorResponse: The error response to send in JSON mode
//
// Returns:
//   - nonce: The generated nonce bytes if successful
//   - error: An error if nonce generation fails
func generateNonceOrFail[T any](
	c cipher.AEAD, w http.ResponseWriter,
	streamModeActive bool, errorResponse T,
) ([]byte, error) {
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		if streamModeActive {
			http.Error(
				w, "failed to generate nonce",
				http.StatusInternalServerError,
			)
			return nil, fmt.Errorf("failed to generate nonce: %w", err)
		}

		return nil, net.Fail(
			errorResponse,
			w,
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate nonce: %w", err),
			"",
		)
	}

	return nonce, nil
}

// encryptData generates a nonce, performs the actual encryption operation,
// and returns the nonce and ciphertext.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - c: The cipher to use for encryption
//   - w: The HTTP response writer for error responses
//   - streamModeActive: Whether the request is in streaming mode
//   - fName: The function name for logging
//
// Returns:
//   - nonce: The generated nonce bytes
//   - ciphertext: The encrypted data
//   - error: An error if nonce generation fails
func encryptData(
	plaintext []byte, c cipher.AEAD, w http.ResponseWriter,
	streamModeActive bool, fName string,
) ([]byte, []byte, error) {
	nonce, err := generateNonceOrFail(
		c, w, streamModeActive,
		reqres.CipherEncryptResponse{Err: data.ErrInternal},
	)
	if err != nil {
		return nil, nil, err
	}

	log.Log().Info(
		fName,
		"message",
		fmt.Sprintf("encrypt %d %d", len(nonce), len(plaintext)),
	)
	ciphertext := c.Seal(nil, nonce, plaintext, nil)

	return nonce, ciphertext, nil
}
