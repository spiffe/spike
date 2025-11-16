//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"

	"github.com/spiffe/spike/internal/journal"
)

// RouteEncrypt handles HTTP requests to encrypt plaintext data using the
// SPIKE Nexus's cipher. This endpoint provides encryption-as-a-service
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
// Errors:
//   - Returns ErrReadFailure if the request body cannot be read
//   - Returns ErrParseFailure if JSON request cannot be parsed
//   - Returns ErrInternal if cipher is unavailable or nonce generation fails
//   - Returns appropriate HTTP status codes for different error conditions
func RouteEncrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "RouteEncrypt"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if streaming mode based on Content-Type
	contentType := r.Header.Get(headerKeyContentType)
	streamModeActive := contentType == headerValueOctetStream

	// Create a cipher getter function (called only after auth passes)
	getCipher := func() (cipher.AEAD, error) {
		return getCipherOrFail(
			w, streamModeActive,
			reqres.CipherEncryptResponse{Err: data.ErrInternal},
		)
	}

	if streamModeActive {
		return handleStreamingEncrypt(w, r, getCipher, fName)
	}

	return handleJSONEncrypt(w, r, getCipher, fName)
}
