//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// guardPolicyReadRequest validates a policy read request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the policy ID format
//   - Checks if the peer has read permission for the policy access path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy read request containing the policy ID
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: nil if all validations pass,
//     ErrAccessUnauthorized if authentication or authorization fails,
//     ErrDataInvalidInput if policy ID validation fails
func guardPolicyReadRequest(
	request reqres.PolicyReadRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := net.ExtractPeerSPIFFEIDFromRequestAndRespondOnFail[reqres.PolicyReadResponse](
		r, w, reqres.PolicyReadResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	policyID := request.ID

	validationErr := validation.ValidatePolicyID(policyID)
	if validationErr != nil {
		failErr := net.Fail(
			reqres.PolicyReadResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return validationErr.Wrap(failErr)
		}
		validationErr.Msg = "invalid policy ID: " + policyID
		return validationErr
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), apiAuth.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		failErr := net.Fail(
			reqres.PolicyReadResponse{}.Unauthorized(), w, http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized.Clone()
	}

	return nil
}
