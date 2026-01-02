//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike-sdk-go/journal"
)

// RouteEncrypt handles HTTP requests to encrypt plaintext data using the
// SPIKE Nexus cipher. This endpoint provides encryption-as-a-service
// functionality without persisting any data.
//
// The function supports two modes based on Content-Type:
//
// 1. Streaming mode (Content-Type: application/octet-stream):
//   - Input: raw binary data to encrypt
//   - Output: version byte + nonce + ciphertext (binary stream)
//
// 2. JSON mode (any other Content-Type):
//   - Input: JSON with { plaintext: []byte, algorithm: string (optional) }
//   - Output: JSON with { version: byte, nonce: []byte,
//     ciphertext: []byte, err: string }
//
// The encryption process:
//  1. Determines mode based on Content-Type header
//  2. For JSON mode: validates request and checks permissions
//  3. Retrieves the system cipher from the backend
//  4. Generates a cryptographically secure random nonce
//  5. Encrypts the data using authenticated encryption (AEAD)
//  6. Returns the encrypted data in the appropriate format
//
// Access control is enforced through guardEncryptSecretRequest for JSON mode.
// Streaming mode may have different permission requirements.
//
// Parameters:
//   - w: HTTP response writer for sending the encrypted response
//   - r: HTTP request containing plaintext data to encrypt
//   - audit: Audit entry for logging the encryption request
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or one of:
//   - ErrDataReadFailure if request body cannot be read
//   - ErrDataParseFailure if JSON request cannot be parsed
//   - ErrStateBackendNotReady if cipher is unavailable
//   - ErrCryptoNonceGenerationFailed if nonce generation fails
func RouteEncrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RouteEncrypt"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if streaming mode based on Content-Type
	contentType := r.Header.Get(headerKeyContentType)
	streamModeActive := contentType == headerValueOctetStream

	if streamModeActive {
		// Cipher getter for streaming mode
		getCipher := func() (cipher.AEAD, *sdkErrors.SDKError) {
			return getCipherOrFailStreaming(w)
		}
		return handleStreamingEncrypt(w, r, getCipher)
	}

	// Cipher getter for JSON mode
	getCipher := func() (cipher.AEAD, *sdkErrors.SDKError) {
		return getCipherOrFailJSON(
			w, reqres.CipherEncryptResponse{Err: sdkErrors.ErrAPIInternal.Code},
		)
	}
	return handleJSONEncrypt(w, r, getCipher)
}
