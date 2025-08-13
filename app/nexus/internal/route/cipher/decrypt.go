//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
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

// RouteDecrypt handles HTTP requests to decrypt ciphertext data using the
// SPIKE Nexus's cipher. This endpoint provides decryption-as-a-service
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
// Errors:
//   - Returns ErrReadFailure if request body cannot be read
//   - Returns ErrParseFailure if JSON request cannot be parsed
//   - Returns ErrBadInput if version is not supported or nonce size is invalid
//   - Returns ErrInternal if cipher is unavailable or decryption fails
//   - Returns appropriate HTTP status codes for different error conditions
func RouteDecrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeDecrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if streaming mode based on Content-Type
	contentType := r.Header.Get("Content-Type")
	streamModeActive := contentType == "application/octet-stream"

	// Get cipher early as both modes need it
	c := persist.Backend().GetCipher()
	if c == nil {
		if streamModeActive {
			http.Error(w, "cipher not available", http.StatusInternalServerError)
			return fmt.Errorf("cipher not available")
		}
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return fmt.Errorf("cipher not available")
	}

	var version byte
	var nonce []byte
	var ciphertext []byte

	if streamModeActive {
		err := guardDecryptSecretRequest(reqres.SecretDecryptRequest{}, w, r)
		if err != nil {
			return err
		}

		// Streaming mode - read the version, nonce, then ciphertext
		ver := make([]byte, 1)
		n, err := io.ReadFull(r.Body, ver)
		if err != nil || n != 1 {
			log.Log().Debug(fName, "message", "Failed to read version")
			http.Error(w, "failed to read version", http.StatusBadRequest)
			return fmt.Errorf("failed to read version")
		}
		version = ver[0]

		// Read nonce
		bytesToRead := c.NonceSize()
		nonce = make([]byte, bytesToRead)
		n, err = io.ReadFull(r.Body, nonce)
		if err != nil || n != bytesToRead {
			log.Log().Debug(fName, "message", "Failed to read nonce")
			http.Error(w, "failed to read nonce", http.StatusBadRequest)
			return fmt.Errorf("failed to read nonce")
		}

		// Read the remaining body as ciphertext
		ciphertext = net.ReadRequestBody(w, r)
		if ciphertext == nil {
			return apiErr.ErrReadFailure
		}
	} else {
		// JSON mode - parse request
		requestBody := net.ReadRequestBody(w, r)
		if requestBody == nil {
			return apiErr.ErrReadFailure
		}

		request := net.HandleRequest[
			reqres.SecretDecryptRequest, reqres.SecretDecryptResponse](
			requestBody, w,
			reqres.SecretDecryptResponse{Err: data.ErrBadInput},
		)
		if request == nil {
			return apiErr.ErrParseFailure
		}

		err := guardDecryptSecretRequest(*request, w, r)
		if err != nil {
			return err
		}

		version = request.Version
		nonce = request.Nonce
		ciphertext = request.Ciphertext
	}

	// Validate version
	if version != byte('1') {
		if streamModeActive {
			http.Error(w, "unsupported version", http.StatusBadRequest)
			return fmt.Errorf("unsupported version: %v", version)
		}
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("unsupported version: %v", version)
	}

	// Validate nonce size
	if len(nonce) != c.NonceSize() {
		if streamModeActive {
			http.Error(w, "invalid nonce size", http.StatusBadRequest)
			return fmt.Errorf("invalid nonce size: expected %d, got %d",
				c.NonceSize(), len(nonce))
		}
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("invalid nonce size: expected %d, got %d",
			c.NonceSize(), len(nonce))
	}

	// Decrypt the ciphertext
	log.Log().Info(fName, "message",
		fmt.Sprintf("Decrypt %d %d", len(nonce), len(ciphertext)),
	)

	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Log().Info(fName, "message", fmt.Errorf("failed to decrypt %w", err))
		if streamModeActive {
			http.Error(w, "decryption failed", http.StatusBadRequest)
			return fmt.Errorf("decryption failed: %w", err)
		}
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("decryption failed: %w", err)
	}

	if streamModeActive {
		// Streaming response: raw plaintext
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(plaintext); err != nil {
			return err
		}
		log.Log().Info(fName, "message", "Streaming decryption successful")
		return nil
	}

	// JSON response
	responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
		Plaintext: plaintext,
		Err:       data.ErrSuccess,
	}, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
