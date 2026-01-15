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

// guardPolicyPutRequest validates a policy creation request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer has write permission for the policy access path
//   - Validates the policy name format
//   - Validates the SPIFFE ID pattern (regex)
//   - Validates the path pattern (regex)
//   - Validates the permissions list
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy creation request containing policy details
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if any input validation fails
func guardPolicyPutRequest(
	request reqres.PolicyPutRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	if authErr := net.AuthorizeAndRespondOnFail(
		reqres.PolicyPutResponse{}.Unauthorized(),
		predicate.AllowSPIFFEIDForPolicyWrite,
		state.CheckAccess,
		w, r,
	); authErr != nil {
		return authErr
	}

	if nameErr := net.RespondErrOnBadName(
		request.Name, reqres.PolicyPutResponse{}.BadRequest(), w,
	); nameErr != nil {
		return nameErr
	}

	if spifeIdPatternErr := net.RespondErrOnBadSPIFFEIDPattern(
		request.SPIFFEIDPattern, reqres.PolicyPutResponse{}.BadRequest(), w,
	); spifeIdPatternErr != nil {
		return spifeIdPatternErr
	}

	if pathPatternErr := net.RespondErrOnBadPathPattern(
		request.PathPattern, reqres.PolicyPutResponse{}.BadRequest(), w,
	); pathPatternErr != nil {
		return pathPatternErr
	}

	return net.RespondErrOnBadPermission(
		request.Permissions, reqres.PolicyPutResponse{}.BadRequest(), w,
	)
}
