//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/journal"
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
) *sdkErrors.SDKError {
	const fName = "routeVerify"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	request, err := net.ReadParseAndGuard[
		reqres.BootstrapVerifyRequest, reqres.BootstrapVerifyResponse](
		w, r, reqres.BootstrapVerifyResponse{}.BadRequest(), guardVerifyRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Get cipher from the backend
	c := persist.Backend().GetCipher()
	if c == nil {
		return net.RespondWithInternalError(
			sdkErrors.ErrCryptoCipherNotAvailable, w,
			reqres.BootstrapVerifyResponse{},
		)
	}

	// Decrypt the ciphertext
	plaintext, decryptErr := c.Open(nil, request.Nonce, request.Ciphertext, nil)
	if decryptErr != nil {
		return net.RespondWithInternalError(
			sdkErrors.ErrCryptoDecryptionFailed, w,
			reqres.BootstrapVerifyResponse{},
		)
	}

	// Compute SHA-256 hash of plaintext
	hash := sha256.Sum256(plaintext)
	hashHex := hex.EncodeToString(hash[:])

	return net.Success(
		reqres.BootstrapVerifyResponse{
			Hash: hashHex,
		}.Success(), w,
	)
}
