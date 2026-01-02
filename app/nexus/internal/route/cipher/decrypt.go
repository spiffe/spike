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

// RouteDecrypt handles HTTP requests to decrypt ciphertext data using the
// SPIKE Nexus cipher. This endpoint provides decryption-as-a-service
// functionality without persisting any data.
//
// The function supports two modes based on Content-Type:
//
// 1. Streaming mode (Content-Type: application/octet-stream):
//   - Input: version byte + nonce + ciphertext (binary stream)
//   - Output: raw decrypted binary data
//
// 2. JSON mode (any other Content-Type):
//   - Input: JSON with { version: byte, nonce: []byte,
//     ciphertext: []byte, algorithm: string (optional) }
//   - Output: JSON with { plaintext: []byte, err: string }
//
// The decryption process:
//  1. Determines mode based on Content-Type header
//  2. For JSON mode: validates request and checks permissions
//  3. Retrieves the system cipher from the backend
//  4. Validates the protocol version and nonce size
//  5. Decrypts the ciphertext using authenticated decryption (AEAD)
//  6. Returns the decrypted plaintext in the appropriate format
//
// Access control is enforced through guardDecryptSecretRequest for JSON mode.
// Streaming mode may have different permission requirements.
//
// Parameters:
//   - w: HTTP response writer for sending the decrypted response
//   - r: HTTP request containing ciphertext data to decrypt
//   - audit: Audit entry for logging the decryption request
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or one of:
//   - ErrDataReadFailure if request body cannot be read
//   - ErrDataParseFailure if JSON request cannot be parsed
//   - ErrDataInvalidInput if the version is unsupported or nonce size is
//     invalid
//   - ErrStateBackendNotReady if cipher is unavailable
//   - ErrCryptoDecryptFailed if decryption fails
func RouteDecrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeDecrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if streaming mode based on Content-Type
	contentType := r.Header.Get(headerKeyContentType)
	streamModeActive := contentType == headerValueOctetStream

	if streamModeActive {
		// Cipher getter for streaming mode
		getCipher := func() (cipher.AEAD, *sdkErrors.SDKError) {
			return getCipherOrFailStreaming(w)
		}
		return handleStreamingDecrypt(w, r, getCipher)
	}

	// Cipher getter for JSON mode
	getCipher := func() (cipher.AEAD, *sdkErrors.SDKError) {
		return getCipherOrFailJSON(
			w, reqres.CipherDecryptResponse{Err: sdkErrors.ErrAPIInternal.Code},
		)
	}
	return handleJSONDecrypt(w, r, getCipher)
}
