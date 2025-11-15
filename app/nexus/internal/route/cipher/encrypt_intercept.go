//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/config/auth"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

// guardEncryptCipherRequest validates a cipher encryption request by
// performing authentication and authorization checks.
//
// This function implements a two-tier authorization model:
//  1. Lite workloads are automatically granted encryption access
//  2. Other workloads must have execute permission for the cipher encrypt path
//
// The function performs the following validations in order:
//   - Extracts the SPIFFE ID from the request
//   - Verifies the SPIFFE ID is not nil
//   - Validates the SPIFFE ID format
//   - Checks if the peer is a lite workload (automatically allowed)
//   - If not a lite workload, checks if the peer has execute permission for
//     the system cipher encrypt path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The cipher encryption request (currently unused, reserved for
//     future validation needs)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
func guardEncryptCipherRequest(
	_ reqres.CipherEncryptRequest, w http.ResponseWriter, r *http.Request,
) error {
	sid, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherEncryptResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded := err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	if sid == nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherEncryptResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded := err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(sid.String())
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherEncryptResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded := err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	// Lite Workloads are always allowed:
	allowed := false
	if spiffeid.IsLiteWorkload(sid.String()) {
		allowed = true
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		allowed = state.CheckAccess(
			sid.String(),
			auth.PathSystemCipherEncrypt,
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}
	// If not, block the request:
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherEncryptResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded := err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	return nil
}
