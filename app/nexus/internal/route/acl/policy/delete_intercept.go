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
//   - Validates the policy name format
//   - Checks if the peer has write permission for the policy access path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy deletion request containing the policy name
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: nil if all validations pass,
//     ErrAccessUnauthorized if authentication or authorization fails,
//     ErrDataInvalidInput if policy name validation fails
func guardPolicyDeleteRequest(
	request reqres.PolicyDeleteRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	if authErr := net.AuthorizeAndRespondOnFail(
		reqres.PolicyDeleteResponse{}.Unauthorized(),
		predicate.AllowSPIFFEIDForPolicyDelete,
		state.CheckPolicyAccess,
		w, r,
	); authErr != nil {
		return authErr
	}

	// SPIKE policies are keyed by name; the SDK's PolicyDeleteRequest
	// still exposes that identifier as ID, so request.ID holds the policy
	// name and must be validated as a name, not as a UUID. See issue #250
	// for the pending SDK field rename.
	return net.RespondErrOnBadName(
		request.ID, reqres.PolicyDeleteResponse{}.BadRequest(), w,
	)
}
