//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"
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
	request reqres.BootstrapVerifyRequest,
	w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	// No CheckAccess because this route is privileged and should not honor
	// policy overrides. Match exact SPIFFE ID instead.
	if authErr := net.AuthorizeAndRespondOnFailNoPolicy(
		reqres.BootstrapVerifyResponse{}.Unauthorized(),
		spiffeid.IsBootstrap,
		w, r,
	); authErr != nil {
		return authErr
	}

	nonceErr := net.RespondCryptoErrOnInvalidNonceSize(
		request.Nonce, w, reqres.BootstrapVerifyResponse{}.BadRequest(),
	)
	if nonceErr != nil {
		return nonceErr
	}

	return net.RespondCryptoErrOnLargeCipherText(
		request.Ciphertext, w, reqres.BootstrapVerifyResponse{}.BadRequest(),
	)
}
