//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	sdkSpiffeid "github.com/spiffe/spike-sdk-go/spiffeid"

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
	request reqres.CipherDecryptRequest,
	peerSPIFFEID *spiffeid.ID,
	w http.ResponseWriter,
	_ *http.Request,
) *sdkErrors.SDKError {
	// Validate version
	if err := validateVersion(
		request.Version, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	// Validate nonce size
	if err := validateNonceSize(
		request.Nonce, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	// Validate ciphertext size to prevent DoS attacks
	if err := validateCiphertextSize(
		request.Ciphertext, w, reqres.CipherDecryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

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
		failErr := net.Fail(
			reqres.CipherDecryptResponse{}.Unauthorized(), w, http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized.Clone()
	}

	return nil
}
