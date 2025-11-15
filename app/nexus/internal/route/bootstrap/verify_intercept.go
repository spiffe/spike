//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/auth"

	"github.com/spiffe/spike/internal/net"
)

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
//   - apiErr.ErrUnauthorized if authentication fails or peer is not bootstrap
//   - apiErr.ErrInvalidInput if nonce or ciphertext validation fails
func guardVerifyRequest(
	request reqres.BootstrapVerifyRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.BootstrapVerifyResponse](
		r, w, reqres.BootstrapVerifyResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	if !spiffeid.IsBootstrap(peerSPIFFEID.String()) {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	// Validate nonce size (AES-GCM standard nonce is 12 bytes)
	const expectedNonceSize = 12 // TODO: to constants.
	if len(request.Nonce) != expectedNonceSize {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	// Validate ciphertext size to prevent DoS attacks
	const maxCiphertextSize = 1024 // TODO: to constants.
	if len(request.Ciphertext) > maxCiphertextSize {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	return nil
}
