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

// guardEncryptCipherRequest validates a cipher encryption request by
// performing authentication, authorization, and request field validation.
//
// This function implements a two-tier authorization model:
//  1. Lite workloads are automatically granted encryption access
//  2. Other workloads must have execute permission for the cipher encrypt path
//
// The function performs the following validations in order:
//   - Validates request fields (future: size limits, format checks, etc.)
//   - Checks if the peer is a lite workload (automatically allowed)
//   - If not a lite workload, checks if the peer has execute permission for
//     the system cipher encrypt path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The cipher encryption request to validate
//   - peerSPIFFEID: The already-validated peer SPIFFE ID (pointer)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request (for context)
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authorization fails
//   - apiErr.ErrBadInput if request validation fails
func guardEncryptCipherRequest(
	request reqres.CipherEncryptRequest,
	peerSPIFFEID *spiffeid.ID,
	w http.ResponseWriter,
	_ *http.Request,
) *sdkErrors.SDKError {
	// Validate plaintext size to prevent DoS attacks
	if err := validatePlaintextSize(
		request.Plaintext, w, reqres.CipherEncryptResponse{}.BadRequest(),
	); err != nil {
		return err
	}

	// Lite Workloads are always allowed:
	allowed := false
	if sdkSpiffeid.IsLiteWorkload(peerSPIFFEID.String()) {
		allowed = true
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		allowed = state.CheckAccess(
			peerSPIFFEID.String(),
			apiAuth.PathSystemCipherEncrypt,
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}
	// If not, block the request:
	if !allowed {
		failErr := net.Fail(
			reqres.CipherEncryptResponse{}.Unauthorized(), w, http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized.Clone()
	}

	return nil
}
