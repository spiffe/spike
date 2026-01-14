//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// guardDecryptCipherRequest validates a cipher decryption request by
// performing authentication, authorization, and request field validation.
//
// This function implements a two-tier authorization model:
//  1. Lite workloads are automatically granted decryption access
//  2. Other workloads must have execute permission for the cipher decrypt path
//
// The function performs the following validations in order:
//   - Validates request fields (future: size limits, format checks, etc.)
//   - Checks if the peer is a lite workload (automatically allowed)
//   - If not a lite workload, checks if the peer has execute permission for
//     the system cipher decrypt path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The cipher decryption request to validate
//   - peerSPIFFEID: The already-validated peer SPIFFE ID (pointer)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request (for context)
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authorization fails
//   - apiErr.ErrBadInput if request validation fails
func guardDecryptCipherRequest(
	request reqres.CipherDecryptRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	// validate peer SPIFFE ID
	_, err := net.ExtractPeerSPIFFEIDAndRespondOnFail(
		w, r, reqres.CipherDecryptResponse{
			Err: sdkErrors.ErrAccessUnauthorized.Code,
		})
	if err != nil {
		return err
	}

	// Validate version
	if err := net.RespondCryptoErrOnVersionMismatch(
		request.Version, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	// Validate nonce size
	if err := net.RespondCryptoErrOnInvalidNonceSize(
		request.Nonce, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	// Validate ciphertext size to prevent DoS attacks
	if err := net.RespondCryptoErrOnLargeCipherText(
		request.Ciphertext, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	return net.RespondUnauthorizedOnPredicateFail(
		func(peerSPIFFEID string) bool {
			return predicate.AllowSPIFFEIDForCipherDecrypt(
				peerSPIFFEID, state.CheckAccess,
			)
		},
		reqres.CipherDecryptResponse{}.Unauthorized(),
		w, r,
	)
}
