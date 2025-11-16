//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	sdkSpiffeid "github.com/spiffe/spike-sdk-go/spiffeid"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// extractAndValidateSPIFFEID extracts and validates the peer SPIFFE ID from
// the request without performing authorization checks. This is used as the
// first step before accessing sensitive resources like the cipher.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *spiffeid.ID: The validated peer SPIFFE ID (pointer)
//   - error: An error if extraction or validation fails
func extractAndValidateSPIFFEID(
	w http.ResponseWriter, r *http.Request,
) (*spiffeid.ID, error) {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.CipherDecryptResponse](
		r, w, reqres.CipherDecryptResponse{
			Err: data.ErrUnauthorized,
		})
	if alreadyResponded := err != nil; alreadyResponded {
		return nil, err
	}

	return peerSPIFFEID, nil
}

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
	request reqres.CipherDecryptRequest,
	peerSPIFFEID *spiffeid.ID,
	w http.ResponseWriter,
	r *http.Request,
) error {
	const fName = "guardDecryptCipherRequest"

	// TODO: Add request field validation here
	// For example: validate ciphertext size limits, nonce format, etc.
	_ = request // Will be used for validation

	// Lite workloads are always allowed:
	allowed := false
	if sdkSpiffeid.IsLiteWorkload(peerSPIFFEID.String()) {
		allowed = true
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		allowed = state.CheckAccess(
			peerSPIFFEID.String(),
			apiAuth.PathSystemCipherDecrypt,
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}

	if !allowed {
		return net.Fail(
			reqres.CipherDecryptResponse{Err: data.ErrUnauthorized},
			w, http.StatusUnauthorized, apiErr.ErrUnauthorized, fName,
		)
	}

	return nil
}
