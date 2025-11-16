//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	stdErrors "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
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
//   - error: nil on successful retrieval or policy not found, error on system
//     failures
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
) error {
	const fName = "RouteGetPolicy"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyReadRequest, reqres.PolicyReadResponse](
		w, r, reqres.PolicyReadBadInput, guardPolicyReadRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	policyID := request.ID

	policy, err := state.GetPolicy(policyID)
	policyFound := err == nil

	internalError := err != nil && !stdErrors.Is(err, state.ErrPolicyNotFound)
	if internalError {
		failErr := stdErrors.Join(errors.ErrQueryFailure, err)
		return net.Fail(
			reqres.PolicyReadResponse{Err: data.ErrInternal}, w,
			http.StatusInternalServerError, failErr, fName,
		)
	}

	if policyFound {
		log.Log().Info(fName, "message", data.ErrFound, "id", policy.ID)
	} else {
		log.Log().Info(fName, "message", data.ErrNotFound,
			"id", policyID, "err", err.Error(),
		)
	}

	if !policyFound {
		return net.Fail(
			reqres.PolicyReadNotFound, w,
			http.StatusNotFound, state.ErrPolicyNotFound, fName,
		)
	}

	net.Success(reqres.PolicyReadResponse{Policy: policy}, w, fName)
	return nil
}
