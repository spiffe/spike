//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/strings"

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
	const fName = "RouteDeletePolicy"
	journal.AuditRequest(fName, r, audit, journal.AuditDelete)
	request, err := net.ReadParseAndGuard[
		reqres.PolicyDeleteRequest, reqres.PolicyDeleteResponse,
	](
		w, r, reqres.PolicyDeleteBadInput, guardPolicyDeleteRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	policyID := request.ID

	err = state.DeletePolicy(policyID)
	if err != nil {
		log.Log().Error(
			fName, "message", data.ErrDeletionFailed, "err", err.Error(),
		)
		return net.Fail(
			reqres.PolicyDeleteInternal, w,
			http.StatusInternalServerError, err,
		)
	}

	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.PolicyDeleteSuccess, w,
	)
	if err == nil {
		net.Respond(http.StatusOK, responseBody, w)
	}
	log.Log().Info(
		fName, "message", data.ErrSuccess, "err", strings.MaybeError(err),
	)
	return nil
}
