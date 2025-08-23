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
	journal "github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteListPolicies handles HTTP requests to retrieve policies.
// It can list all policies or filter them by a SPIFFE ID pattern or a path
// pattern. The function returns a list of policies matching the criteria.
//
// The request body can be empty to list all policies, or it can contain
// `spiffe_id_pattern` or `path_pattern` for filtering. These two filter
// parameters cannot be used together.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request for the policy listing operation
//   - audit: Audit entry for logging the policy list action
//
// Returns:
//   - error: nil on successful retrieval, error otherwise
//
// Example request body (list all):
//
//	{}
//
// Example request body (filter by SPIFFE ID):
//
//	{
//	    "spiffe_id_pattern": "spiffe://example\.org/app"
//	}
//
// Example request body (filter by path):
//
//	{
//	    "path_pattern": "^api/v1/"
//	}
//
// Possible errors:
//   - Failed to read request body
//   - Failed to parse request body
//   - `spiffe_id_pattern` and `path_pattern` used together
//   - Failed to marshal response body
func RouteListPolicies(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	fName := "routeListPolicies"
	journal.AuditRequest(fName, r, audit, journal.AuditList)

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

	var policies []data.Policy

	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern

	switch {
	case SPIFFEIDPattern != "":
		policies, err = state.ListPoliciesBySPIFFEID(SPIFFEIDPattern)
		if err != nil {
			return err
		}
	case pathPattern != "":
		policies, err = state.ListPoliciesByPath(pathPattern)
		if err != nil {
			return err
		}
	default:
		policies, err = state.ListPolicies()
		if err != nil {
			return err
		}
	}

	responseBody := net.MarshalBody(reqres.PolicyListResponse{
		Policies: policies,
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
