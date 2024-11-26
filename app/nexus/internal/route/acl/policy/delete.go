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

// RouteDeletePolicy handles HTTP DELETE requests to remove existing policies.
// It processes the request body to delete a policy specified by its ID.
//
// The function expects a JSON request body containing:
//   - Id: unique identifier of the policy to delete
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
//   - error: nil on successful policy deletion, error otherwise
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
//   - Failed to delete policy (internal server error)
func RouteDeletePolicy(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeDeletePolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditDelete

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyDeleteRequest, reqres.PolicyDeleteResponse](
		requestBody, w,
		reqres.PolicyDeleteResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	policyId := request.Id

	err := state.DeletePolicy(policyId)
	if err != nil {
		log.Log().Info("routeDeletePolicy",
			"msg", "Failed to delete policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: "Internal server error",
		}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeDeletePolicy", "msg", "internal server error")
		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeDeletePolicy", "msg", "OK")
	return nil
}
