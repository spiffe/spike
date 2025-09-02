//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
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
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeDeletePolicy"
	journal.AuditRequest(fName, r, audit, journal.AuditDelete)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.PolicyDeleteRequest, reqres.PolicyDeleteResponse](
		requestBody, w,
		reqres.PolicyDeleteResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	policyID := request.ID

	err := guardDeletePolicyRequest(*request, w, r)
	if err != nil {
		return err
	}

	err = state.DeletePolicy(policyID)
	if err != nil {
		log.Log().Warn(fName, "message", "Failed to delete policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Error(fName, "message", data.ErrInternal)
		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
