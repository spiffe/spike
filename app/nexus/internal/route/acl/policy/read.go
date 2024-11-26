//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// TODO: the request and response can be generic and part of Route's signature.

// RouteGetPolicy handles HTTP requests to retrieve a specific policy by its ID.
// It processes the request body to fetch detailed information about a single
// policy.
//
// The function expects a JSON request body containing:
//   - Id: unique identifier of the policy to retrieve
//
// On success, it returns the complete policy object. If the policy is not found,
// it returns a not found error. For other errors, it returns an internal
// server error.
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
//	        "spiffe_id_pattern": "spiffe://example.org/*/service",
//	        "path_pattern": "/api/*",
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
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeGetPolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.Query())
	audit.Action = log.AuditRead

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyReadRequest, reqres.PolicyReadResponse](
		requestBody, w,
		reqres.PolicyReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	policyId := request.Id

	policy, err := state.GetPolicy(policyId)
	if err == nil {
		log.Log().Info("routeGetPolicy", "msg", "Policy found")
	} else if errors.Is(err, state.ErrPolicyNotFound) {
		log.Log().Info("routeGetPolicy", "msg", "Policy not found")

		res := reqres.PolicyReadResponse{Err: reqres.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		log.Log().Info("routeGetPolicy", "msg", "not found")
		return nil
	} else {
		log.Log().Info("routeGetPolicy",
			"msg", "Failed to retrieve policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyReadResponse{
			Err: "Internal server error"}, w,
		)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeGetPolicy", "msg", "internal server error")
		return err
	}

	responseBody := net.MarshalBody(
		reqres.PolicyReadResponse{Policy: policy}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetPolicy", "msg", "OK")

	return nil
}
