//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	journal "github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RoutePutPolicy handles HTTP PUT requests for creating new policies.
// It processes the request body to create a policy with the specified name,
// SPIFFE ID pattern, path pattern, and permissions.
//
// The function expects a JSON request body containing:
//   - Name: policy name
//   - SpiffeIdPattern: SPIFFE ID matching pattern
//   - PathPattern: path matching pattern
//   - Permissions: set of allowed permissions
//
// On success, it returns a JSON response with the created policy's ID.
// On failure, it returns an appropriate error response with status code.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy creation data
//   - audit: Audit entry for logging the policy creation action
//
// Returns:
//   - error: nil on successful policy creation, error otherwise
//
// Example request body:
//
//	{
//	    "name": "example-policy",
//	    "spiffe_id_pattern": "spiffe://example\.org/*/service",
//	    "path_pattern": "/api/*",
//	    "permissions": ["read", "write"]
//	}
//
// Example success response:
//
//	{
//	    "id": "policy-123"
//	}
//
// Example error response:
//
//	{
//	    "err": "Internal server error"
//	}
func RoutePutPolicy(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routePutPolicy"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrParseFailure
	}

	request := net.HandleRequest[
		reqres.PolicyCreateRequest, reqres.PolicyCreateResponse](
		requestBody, w,
		reqres.PolicyCreateResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrReadFailure
	}

	err := guardPutPolicyRequest(*request, w, r)
	if err != nil {
		return err
	}

	name := request.Name
	spiffeIdPattern := request.SpiffeIdPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	policy, err := state.CreatePolicy(data.Policy{
		Id:              "",
		Name:            name,
		SpiffeIdPattern: spiffeIdPattern,
		PathPattern:     pathPattern,
		Permissions:     permissions,
		CreatedAt:       time.Time{},
		CreatedBy:       "",
	})
	if err != nil {
		log.Log().Warn(fName, "message", "Failed to create policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrInternal,
		}, w)

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Error(fName, "message", data.ErrInternal)

		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
		Id: policy.Id,
	}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)

	return nil
}
