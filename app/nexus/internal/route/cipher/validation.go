//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"fmt"
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/internal/net"
)

// validateVersion validates that the protocol version is supported.
//
// Parameters:
//   - version: The protocol version byte to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//   - fName: The function name for logging
//
// Returns:
//   - nil if the version is valid
//   - error if the version is unsupported
func validateVersion[T any](
	version byte, w http.ResponseWriter, errorResponse T, fName string,
) error {
	if version != spikeCipherVersion {
		net.Fail(errorResponse, w, http.StatusBadRequest)
		return fmt.Errorf("unsupported version: %v", version)
	}
	return nil
}

// validateNonceSize validates that the nonce is exactly the expected size.
//
// Parameters:
//   - nonce: The nonce bytes to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//   - fName: The function name for logging
//
// Returns:
//   - nil if the nonce size is valid
//   - error if the nonce size is invalid
func validateNonceSize[T any](
	nonce []byte, w http.ResponseWriter, errorResponse T, fName string,
) error {
	if len(nonce) != expectedNonceSize {
		net.Fail(errorResponse, w, http.StatusBadRequest)
		return sdkErrors.ErrInvalidInput
	}
	return nil
}

// validateCiphertextSize validates that the ciphertext does not exceed the
// maximum allowed size.
//
// Parameters:
//   - ciphertext: The ciphertext bytes to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//   - fName: The function name for logging
//
// Returns:
//   - nil if the ciphertext size is valid
//   - error if the ciphertext is too large
func validateCiphertextSize[T any](
	ciphertext []byte, w http.ResponseWriter, errorResponse T, fName string,
) error {
	if len(ciphertext) > maxCiphertextSize {
		net.Fail(errorResponse, w, http.StatusBadRequest)
		return sdkErrors.ErrInvalidInput
	}
	return nil
}

// validatePlaintextSize validates that the plaintext does not exceed the
// maximum allowed size.
//
// Parameters:
//   - plaintext: The plaintext bytes to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//   - fName: The function name for logging
//
// Returns:
//   - nil if the plaintext size is valid
//   - error if the plaintext is too large
func validatePlaintextSize[T any](
	plaintext []byte, w http.ResponseWriter, errorResponse T, fName string,
) error {
	if len(plaintext) > maxPlaintextSize {
		net.Fail(errorResponse, w, http.StatusBadRequest)
		return sdkErrors.ErrInvalidInput
	}
	return nil
}
