//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

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
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
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
	const fName = "routeEncrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if streaming mode based on Content-Type
	contentType := r.Header.Get("Content-Type")
	streamModeActive := contentType == "application/octet-stream"

	// TODO: we should do this AFTER the guard interception; not before.
	// TODO: also extract this part into a function.
	// Get cipher early as both modes need it
	c := persist.Backend().GetCipher()
	if c == nil {
		if streamModeActive {
			http.Error(w, "cipher not available", http.StatusInternalServerError)
			return fmt.Errorf("cipher not available") // TODO: symbolic constant.
		}
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherEncryptResponse{
				Err: data.ErrInternal,
			}, w)
		if err == nil {
			net.Respond(http.StatusInternalServerError, responseBody, w)
		}

		return fmt.Errorf("cipher not available") // TODO: symbolic constant.
	}

	var plaintext []byte // TODO: why var?

	if streamModeActive {
		// TODO: this is called twice; elevate it to an earlier place.
		// TODO: we probably need a stream-specific guard too.
		err := guardEncryptCipherRequest(reqres.CipherEncryptRequest{}, w, r)
		if err != nil {
			return err
		}

		// Streaming mode - read raw body
		requestBody := net.ReadRequestBody(w, r)
		if requestBody == nil {
			return apiErr.ErrReadFailure // TODO: we return an error but don't respond on http stream.
		}

		plaintext = requestBody
	} else {
		// JSON mode - parse request
		requestBody := net.ReadRequestBody(w, r)
		if requestBody == nil {
			return apiErr.ErrReadFailure // TODO: we still need to respond to the client.
		}

		request := net.HandleRequest[
			reqres.CipherEncryptRequest, reqres.CipherEncryptResponse](
			requestBody, w,
			reqres.CipherEncryptResponse{Err: data.ErrBadInput},
		)
		if request == nil {
			return apiErr.ErrParseFailure // TODO: we still need to respond to the client.
		}

		err := guardEncryptCipherRequest(*request, w, r)
		if err != nil {
			return err
		}

		plaintext = request.Plaintext
	}

	// Generate nonce
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		if streamModeActive {
			http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
			return fmt.Errorf("failed to generate nonce: %w", err) // TODO: symbolic constant.
		}

		// JSON response:
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(reqres.CipherEncryptResponse{
			Err: data.ErrInternal,
		}, w)
		if err == nil {
			net.Respond(http.StatusInternalServerError, responseBody, w)
		}

		return fmt.Errorf("failed to generate nonce: %w", err) // TODO: symbolic constant.
	}

	// Encrypt the plaintext
	log.Log().Info(fName,
		"message", fmt.Sprintf("Encrypt %d %d", len(nonce), len(plaintext)))
	ciphertext := c.Seal(nil, nonce, plaintext, nil)
	log.Log().Info(fName,
		"message", fmt.Sprintf("len after %d %d", len(nonce), len(ciphertext)))

	if streamModeActive {
		// Streaming response: version + nonce + ciphertext
		w.Header().Set("Content-Type", "application/octet-stream")
		v := byte('1')
		if _, err := w.Write([]byte{v}); err != nil {
			return err
		}
		if _, err := w.Write(nonce); err != nil {
			return err
		}
		if _, err := w.Write(ciphertext); err != nil {
			return err
		}
		log.Log().Info(fName, "message", "Streaming encryption successful")
		return nil
	}

	// JSON response
	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.CipherEncryptResponse{
			Version:    byte('1'), // TODO: maybe set version as constant.
			Nonce:      nonce,
			Ciphertext: ciphertext,
			Err:        data.ErrSuccess,
		}, w)
	if err == nil {
		net.Respond(http.StatusOK, responseBody, w)
	}

	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
