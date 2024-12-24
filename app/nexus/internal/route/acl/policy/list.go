//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteListPolicies handles HTTP requests to retrieve all existing policies.
// It returns a list of all policies in the system, including their IDs, names,
// SPIFFE ID patterns, path patterns, and permissions.
//
// The function expects an empty JSON request body ({}) and returns an array
// of policy objects.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request for the policy listing operation
//   - audit: Audit entry for logging the policy list action
//
// Returns:
//   - error: nil on successful retrieval, error otherwise
//
// Example request body:
//
//	{}
//
// Example success response:
//
//	{
//	    "policies": [
//	        {
//	            "id": "policy-123",
//	            "name": "example-policy",
//	            "spiffe_id_pattern": "spiffe://example.org/*/service",
//	            "path_pattern": "/api/*",
//	            "permissions": ["read", "write"],
//	            "created_at": "2024-01-01T00:00:00Z",
//	            "created_by": "user-abc"
//	        }
//	        // ... additional policies
//	    ]
//	}
//
// Example error response:
//
//	{
//	    "err": "Internal server error"
//	}
//
// Possible errors:
//   - Failed to read request body
//   - Failed to parse request body
//   - Failed to marshal response body
func RouteListPolicies(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	fName := "routeListPolicies"
	log.AuditRequest(fName, r, audit, log.AuditList)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.PolicyListRequest, reqres.PolicyListResponse](
		requestBody, w,
		reqres.PolicyListResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	err := guardListPolicyRequest(*request, w, r)
	if err != nil {
		return err
	}

	policies := state.ListPolicies()

	responseBody := net.MarshalBody(reqres.PolicyListResponse{
		Policies: policies,
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
