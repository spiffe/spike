//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/spiffe"
)

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
	const fName = "routeGetPolicy"
	log.AuditRequest(fName, r, audit, log.AuditRead)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return apiErr.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.PolicyReadRequest, reqres.PolicyReadResponse](
		requestBody, w,
		reqres.PolicyReadResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return apiErr.ErrParseFailure
	}

	policyId := request.Id

	spiffeId, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}

	allowed := state.CheckAccess(
		spiffeId.String(), "*",
		[]data.PolicyPermission{data.PermissionSuper},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	policy, err := state.GetPolicy(policyId)
	if err == nil {
		log.Log().Info(fName, "msg", "Policy found")
	} else if errors.Is(err, state.ErrPolicyNotFound) {
		log.Log().Info(fName, "msg", "Policy not found")

		res := reqres.PolicyReadResponse{Err: data.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		log.Log().Info(fName, "msg", "not found")
		return nil
	} else {
		// TODO: these can be Err constants too (similar to ErrInternal)
		log.Log().Info(fName, "msg", "Failed to retrieve policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyReadResponse{
			Err: data.ErrInternal}, w,
		)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info(fName, "msg", data.ErrInternal)
		return err
	}

	responseBody := net.MarshalBody(
		reqres.PolicyReadResponse{Policy: policy}, w,
	)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)

	return nil
}
