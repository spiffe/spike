//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/crypto"
)

// expectedNonceSize is the standard AES-GCM nonce size. See ADR-0032.
// (https://spike.ist/architecture/adrs/adr-0032/)
const expectedNonceSize = crypto.GCMNonceSize

// guardVerifyRequest validates a bootstrap verification request by performing
// authentication and input validation checks.
//
// This function ensures that only authorized bootstrap instances can verify
// the system initialization by validating cryptographic parameters and peer
// identity.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Verifies the peer has a bootstrap SPIFFE ID
//   - Validates the nonce size (must be 12 bytes for AES-GCM standard)
//   - Validates the ciphertext size (must not exceed 1024 bytes to prevent DoS
//     attacks)
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The bootstrap verification request containing nonce and
//     ciphertext
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - sdkErrors.ErrAccessUnauthorized if authentication fails or peer is not
//     bootstrap
//   - sdkErrors.ErrDataInvalidInput if nonce or ciphertext validation fails
func guardVerifyRequest(
	request reqres.BootstrapVerifyRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.BootstrapVerifyResponse](
		r, w, reqres.BootstrapVerifyResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	if !spiffeid.IsBootstrap(peerSPIFFEID.String()) {
		failErr := net.Fail(
			reqres.BootstrapVerifyResponse{}.Unauthorized(), w,
			http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized.Clone()
	}

	if len(request.Nonce) != expectedNonceSize {
		failErr := net.Fail(
			reqres.BootstrapVerifyResponse{}.BadRequest(), w,
			http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput.Clone()
	}

	// Limit cipherText size to prevent DoS attacks
	// The maximum possible size is 68,719,476,704
	// The limit comes from GCM's 32-bit counter.
	if len(request.Ciphertext) > env.CryptoMaxCiphertextSizeVal() {
		failErr := net.Fail(
			reqres.BootstrapVerifyResponse{}.BadRequest(), w,
			http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput.Clone()
	}

	return nil
}
