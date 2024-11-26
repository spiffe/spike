//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/pkg/spiffe"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
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
	log.Log().Info("routeListPolicies", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditList

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyListRequest, reqres.PolicyListResponse](
		requestBody, w,
		reqres.PolicyListResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyListResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	allowed := state.CheckAccess(
		spiffeid.String(),
		"*",
		[]data.PolicyPermission{data.PermissionSuper},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.PolicyListResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.New("unauthorized")
	}

	policies := state.ListPolicies()

	responseBody := net.MarshalBody(reqres.PolicyListResponse{
		Policies: policies,
	}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeListPolicies", "msg", "success")
	return nil
}
