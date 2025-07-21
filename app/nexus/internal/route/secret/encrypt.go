//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	journal "github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteEncrypt handles HTTP requests to encrypt plaintext data using the
// system's cipher. This endpoint provides encryption-as-a-service
// functionality without persisting any data.
//
// The function expects a JSON request body containing:
//   - plaintext: raw bytes to encrypt
//   - algorithm: (optional) encryption algorithm to use
//
// On success, it returns a JSON response containing:
//   - version: protocol version byte (currently '1')
//   - nonce: randomly generated nonce used for encryption
//   - ciphertext: the encrypted data
//   - err: error code (ErrSuccess on success)
//
// The encryption process:
//  1. Validates and parses the incoming request
//  2. Checks access permissions via guardEncryptSecretRequest
//  3. Retrieves the system cipher from the backend
//  4. Generates a cryptographically secure random nonce
//  5. Encrypts the plaintext using authenticated encryption (AEAD)
//  6. Returns the encrypted data with metadata
//
// Access control is enforced through guardEncryptSecretRequest, which should
// verify the caller has appropriate permissions to use the encryption service.
//
// Errors:
//   - Returns ErrReadFailure if request body cannot be read
//   - Returns ErrParseFailure if request cannot be parsed as SecretEncryptRequest
//   - Returns ErrInternal if cipher is unavailable or nonce generation fails
//   - Returns ErrMarshalFailure if response cannot be marshaled
//   - May return errors from guardEncryptSecretRequest for permission failures
func RouteEncrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeEncrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return apiErr.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretEncryptRequest, reqres.SecretEncryptResponse](
		requestBody, w,
		reqres.SecretEncryptResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return apiErr.ErrParseFailure
	}

	err := guardEncryptSecretRequest(*request, w, r)
	if err != nil {
		return err
	}

	// Get cipher from the backend:
	c := persist.Backend().GetCipher()
	if c == nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return fmt.Errorf("cipher not available")
	}

	// Generate nonce
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	log.Log().Info(fName,
		"message", fmt.Sprintf("Encrypt %d %d",
			len(nonce), len(request.Plaintext)))
	ciphertext := c.Seal(nil, nonce, request.Plaintext, nil)
	log.Log().Info(fName,
		"message", fmt.Sprintf("len after %d %d",
			len(nonce), len(ciphertext)))

	// Prepare response
	responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
		Version:    byte('1'),
		Nonce:      nonce,
		Ciphertext: ciphertext,
		Err:        data.ErrSuccess,
	}, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
