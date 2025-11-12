//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/app/nexus/internal/state/base"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteVerify handles HTTP requests from SPIKE Bootstrap to verify that
// SPIKE Nexus has been properly initialized and can decrypt data using the
// root key.
//
// This endpoint serves as a verification mechanism during the bootstrap
// process. Bootstrap encrypts a random text with the root key and sends it
// to Nexus. Nexus decrypts the text, computes its SHA-256 hash, and returns
// the hash to Bootstrap. Bootstrap can then verify the hash matches the
// original plaintext, confirming that Nexus has been properly initialized
// with the correct root key.
//
// The verification process:
//  1. Reads and validates the request containing nonce and ciphertext
//  2. Checks that the request comes from a Bootstrap SPIFFE ID
//  3. Retrieves the system cipher from the backend
//  4. Decrypts the ciphertext using the nonce
//  5. Computes SHA-256 hash of the decrypted plaintext
//  6. Returns the hash to Bootstrap for verification
//
// Access control is enforced through guardVerifyRequest to ensure only
// Bootstrap can call this endpoint.
//
// Parameters:
//   - w http.ResponseWriter: The HTTP response writer
//   - r *http.Request: The incoming HTTP request
//   - audit *journal.AuditEntry: Audit entry for logging
//
// Returns:
//   - error: An error if one occurs during processing, nil otherwise
//
// Errors:
//   - Returns ErrReadFailure if request body cannot be read
//   - Returns ErrParseFailure if JSON request cannot be parsed
//   - Returns ErrInternal if cipher is unavailable or decryption fails
//   - Returns ErrUnauthorized if request is not from Bootstrap
func RouteVerify(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeVerify"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		log.Log().Warn(fName, "message", "requestBody is nil")
		return apiErr.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.BootstrapVerifyRequest, reqres.BootstrapVerifyResponse](
		requestBody, w,
		reqres.BootstrapVerifyResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		log.Log().Warn(fName, "message", "request is nil")
		return apiErr.ErrParseFailure
	}

	err := guardVerifyRequest(*request, w, r)
	if err != nil {
		return err
	}

	// Get cipher from the backend
	c := persist.Backend().GetCipher()
	if c == nil {
		log.Log().Error(fName, "message", "cipher not available")
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrInternal,
			}, w)
		if err == nil {
			net.Respond(http.StatusInternalServerError, responseBody, w)
		}
		// TODO: this needs to come from a symbolic constant.
		return fmt.Errorf("cipher not available")
	}

	// Decrypt the ciphertext
	// TODO: these need to be removed!
	fmt.Println("nonce", hex.EncodeToString(request.Nonce))
	fmt.Println("ciphertext", hex.EncodeToString(request.Ciphertext))
	fmt.Println("rootKey", hex.EncodeToString(base.RootKeyNoLock()[:]))
	plaintext, err := c.Open(nil, request.Nonce, request.Ciphertext, nil)
	if err != nil {
		log.Log().Error(fName, "message", "decryption failed", "err",
			err.Error())
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrInternal,
			}, w)
		if err == nil {
			net.Respond(http.StatusInternalServerError, responseBody, w)
		}

		// TODO: needs to come from a symbolic constant.
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Compute SHA-256 hash of plaintext
	hash := sha256.Sum256(plaintext)
	hashHex := hex.EncodeToString(hash[:])

	log.Log().Info(fName, "message", "verification successful",
		"plaintext_len", len(plaintext), "hash", hashHex)

	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.BootstrapVerifyResponse{
			Hash: hashHex,
			Err:  data.ErrSuccess,
		}, w)
	if err == nil {
		net.Respond(http.StatusOK, responseBody, w)
	}

	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
