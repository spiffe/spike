//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// guardPolicyDeleteRequest validates a policy deletion request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the policy ID format
//   - Checks if the peer has write permission for the policy access path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy deletion request containing the policy ID
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: nil if all validations pass,
//     ErrAccessUnauthorized if authentication or authorization fails,
//     ErrDataInvalidInput if policy ID validation fails
func guardPolicyDeleteRequest(
	request reqres.PolicyDeleteRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	policyID := request.ID

	// TODO: ensure this happens in ALL guard calls.
	// Extract and validate SPIFFE ID before any action.
	_, err := net.ExtractPeerSPIFFEIDAndRespondOnFail(
		w, r, reqres.PolicyDeleteResponse{
			Err: sdkErrors.ErrAccessUnauthorized.Code,
		})
	if err != nil {
		return err
	}

	// TODO: ensure policy verification before other verifications on ALL guard calls.
	authErr := net.RespondUnauthorizedOnPredicateFail(
		func(peerSPIFFEID string) bool {
			return predicate.AllowSPIFFEIDForPolicyDelete(
				peerSPIFFEID, state.CheckAccess,
			)
		},
		reqres.PolicyDeleteResponse{}.Unauthorized(), w, r,
	)
	if authErr != nil {
		return authErr
	}

	return net.RespondErrOnBadPolicyID(
		policyID, w, reqres.PolicyDeleteResponse{}.BadRequest(),
	)
}
