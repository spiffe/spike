//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike-sdk-go/journal"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// RouteDeletePolicy handles HTTP DELETE requests to remove existing policies.
// It processes the request body to delete a policy specified by its ID.
//
// The function expects a JSON request body containing:
//   - ID: unique identifier of the policy to delete
//
// On success, it returns an empty JSON response with HTTP 200 status.
// On failure, it returns an appropriate error response with status code.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy ID to delete
//   - audit: Audit entry for logging the policy deletion action
//
// Returns:
//   - *sdkErrors.SDKError: nil on successful policy deletion, error otherwise
//
// Example request body:
//
//	{
//	    "id": "policy-123"
//	}
//
// Example success response:
//
//	{}
//
// Example not found response:
//
//	{
//	    "err": "not_found"
//	}
//
// Example error response:
//
//	{
//	    "err": "Internal server error"
//	}
//
// HTTP Status Codes:
//   - 200: Policy deleted successfully
//   - 404: Policy not found
//   - 500: Internal server error
//
// Possible errors:
//   - Failed to read request body
//   - Failed to parse request body
//   - Policy not found
//   - Internal server error during policy deletion
func RouteDeletePolicy(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RouteDeletePolicy"

	journal.AuditRequest(fName, r, audit, journal.AuditDelete)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyDeleteRequest, reqres.PolicyDeleteResponse,
	](
		w, r, reqres.PolicyDeleteResponse{}.BadRequest(), guardPolicyDeleteRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	policyID := request.ID

	deleteErr := state.DeletePolicy(policyID)
	if deleteErr != nil {
		return net.RespondWithHTTPError(deleteErr, w, reqres.PolicyDeleteResponse{})
	}

	return net.Success(reqres.PolicyDeleteResponse{}.Success(), w)
}
