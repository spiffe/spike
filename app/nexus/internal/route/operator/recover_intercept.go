//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/auth"
)

// guardRecoverRequest validates a system recovery request by performing
// authentication and authorization checks.
//
// This function implements strict authorization for system recovery operations,
// which are critical administrative functions that should only be accessible
// to authorized operator identities.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Verifies the peer has a pilot-recover SPIFFE ID (operator role)
//
// Only identities with the pilot-recover role are authorized to perform system
// recovery operations. All other identities are rejected.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The recovery request (currently unused, reserved for future
//     validation needs)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: An error if authentication fails or the peer is
//     not authorized (not pilot-recover). Returns nil if all validations pass.
func guardRecoverRequest(
	_ reqres.RecoverRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.RestoreResponse](
		r, w, reqres.RestoreResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// We don't do policy checks as the recovery operation purely restricted to
	// SPIKE Pilot.
	if !spiffeid.IsPilotRecover(peerSPIFFEID.String()) {
		failErr := net.Fail(
			reqres.RestoreResponse{}.Unauthorized(), w, http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized
	}

	return nil
}
