//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
)

// RouteGetPolicy handles HTTP requests to retrieve a specific policy by its ID.
// It processes the request body to fetch detailed information about a single
// policy.
//
// The function expects a JSON request body containing:
//   - ID: unique identifier of the policy to retrieve
//
// On success, it returns the complete policy object. If the policy is not
// found, it returns a "not found" error. For other errors, it returns an
// internal server error.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy ID to retrieve
//   - audit: Audit entry for logging the policy read action
//
// Returns:
//   - *sdkErrors.SDKError: nil on successful retrieval, ErrEntityNotFound if
//     policy not found, other errors on system failures
//
// Example request body:
//
//	{
//	    "id": "policy-123"
//	}
//
// Example success response:
//
//	{
//	    "policy": {
//	        "id": "policy-123",
//	        "name": "example-policy",
//	        "spiffe_id_pattern": "^spiffe://example\.org/.*/service",
//	        "path_pattern": "^api/",
//	        "permissions": ["read", "write"],
//	        "created_at": "2024-01-01T00:00:00Z",
//	        "created_by": "user-abc"
//	    }
//	}
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
//   - 200: Policy found and returned successfully
//   - 404: Policy not found
//   - 500: Internal server error
//
// Possible errors:
//   - Failed to read request body
//   - Failed to parse request body
//   - Failed to marshal response body
//   - Policy not found
//   - Internal server error during policy retrieval
func RouteGetPolicy(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RouteGetPolicy"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyReadRequest, reqres.PolicyReadResponse](
		w, r, reqres.PolicyReadResponse{}.BadRequest(), guardPolicyReadRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	policyID := request.ID

	policy, policyErr := state.GetPolicy(policyID)
	if policyErr != nil {
		return net.RespondWithHTTPError(policyErr, w, reqres.PolicyReadResponse{})
	}

	return net.Success(reqres.PolicyReadResponse{Policy: policy}.Success(), w)
}
