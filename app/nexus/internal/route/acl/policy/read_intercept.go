//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	stdErrs "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/internal/auth"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
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
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if policy ID validation fails
func guardPolicyReadRequest(
	request reqres.PolicyReadRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	const fName = "guardPolicyReadRequest"

	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.PolicyReadResponse](
		r, w, reqres.PolicyReadUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	policyID := request.ID

	err = validation.ValidatePolicyID(policyID)
	if err != nil {
		failErr := stdErrs.Join(sdkErrors.ErrInvalidInput, err)
		return net.Fail(
			reqres.PolicyReadBadInput, w,
			http.StatusBadRequest, failErr, fName,
		)
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), apiAuth.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		return net.Fail(
			reqres.PolicyReadUnauthorized, w,
			http.StatusUnauthorized, sdkErrors.ErrUnauthorized, fName,
		)
	}

	return nil
}
