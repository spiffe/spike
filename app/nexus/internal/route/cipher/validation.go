//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
)

// extractAndValidateSPIFFEID extracts and validates the peer SPIFFE ID from
// the request without performing authorization checks. This is used as the
// first step before accessing sensitive resources like the cipher.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *spiffeid.ID: The validated peer SPIFFE ID (pointer)
//   - error: An error if extraction or validation fails
func extractAndValidateSPIFFEID(
	w http.ResponseWriter, r *http.Request,
) (*spiffeid.ID, *sdkErrors.SDKError) {
	peerSPIFFEID, err := net.ExtractPeerSPIFFEIDFromRequestAndRespondOnFail[reqres.CipherDecryptResponse](
		r, w, reqres.CipherDecryptResponse{
			Err: sdkErrors.ErrAccessUnauthorized.Code,
		})
	if alreadyResponded := err != nil; alreadyResponded {
		return nil, err
	}

	return peerSPIFFEID, nil
}

// validateVersion validates that the protocol version is supported.
//
// Parameters:
//   - version: The protocol version byte to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//
// Returns:
//   - nil if the version is valid
//   - *sdkErrors.SDKError if the version is unsupported
func validateVersion[T any](
	version byte, w http.ResponseWriter, errorResponse T,
) *sdkErrors.SDKError {
	if version != spikeCipherVersion {
		failErr := net.Fail(errorResponse, w, http.StatusBadRequest)
		if failErr != nil {
			return sdkErrors.ErrCryptoUnsupportedCipherVersion.Wrap(failErr)
		}
		return sdkErrors.ErrCryptoUnsupportedCipherVersion.Clone()
	}
	return nil
}

// validateNonceSize validates that the nonce is exactly the expected size.
//
// Parameters:
//   - nonce: The nonce bytes to validate
//   - w: The HTTP response writer for error responses
//   - errorResponse: The error response to send on failure
//
// Returns:
//   - nil if the nonce size is valid
//   - *sdkErrors.SDKError if the nonce size is invalid
func validateNonceSize[T any](
	nonce []byte, w http.ResponseWriter, errorResponse T,
) *sdkErrors.SDKError {
	if len(nonce) != expectedNonceSize {
		failErr := net.Fail(errorResponse, w, http.StatusBadRequest)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput.Clone()
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
//
// Returns:
//   - nil if the ciphertext size is valid
//   - *sdkErrors.SDKError if the ciphertext is too large
func validateCiphertextSize[T any](
	ciphertext []byte, w http.ResponseWriter, errorResponse T,
) *sdkErrors.SDKError {
	if len(ciphertext) > env.CryptoMaxCiphertextSizeVal() {
		failErr := net.Fail(errorResponse, w, http.StatusBadRequest)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput.Clone()
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
//
// Returns:
//   - nil if the plaintext size is valid
//   - *sdkErrors.SDKError if the plaintext is too large
func validatePlaintextSize[T any](
	plaintext []byte, w http.ResponseWriter, errorResponse T,
) *sdkErrors.SDKError {
	if len(plaintext) > env.CryptoMaxPlaintextSizeVal() {
		failErr := net.Fail(errorResponse, w, http.StatusBadRequest)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput
	}
	return nil
}
