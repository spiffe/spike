//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
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
//   - *sdkErrors.SDKError: nil on successful retrieval, error otherwise
//
// Example request body (list all):
//
//	{}
//
// Example request body (filter by SPIFFE ID):
//
//	{
//	    "spiffe_id_pattern": "^spiffe://example\\.org/app$"
//	}
//
// Example request body (filter by path):
//
//	{
//	    "path_pattern": "^secrets/db/.*$"
//	}
//
// Possible errors:
//   - Failed to read request body
//   - Failed to parse request body
//   - `spiffe_id_pattern` and `path_pattern` used together (validated by the
//     request guard)
//   - Failed to marshal response body
func RouteListPolicies(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	fName := "RouteListPolicies"

	journal.AuditRequest(fName, r, audit, journal.AuditList)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyListRequest, reqres.PolicyListResponse](
		w, r, reqres.PolicyListResponse{}.BadRequest(), guardListPolicyRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	var policies []data.Policy

	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern

	var listErr *sdkErrors.SDKError

	// Note that Go's default switch behavior will not fall through.
	switch {
	case SPIFFEIDPattern != "":
		policies, listErr = state.ListPoliciesBySPIFFEIDPattern(SPIFFEIDPattern)
	case pathPattern != "":
		policies, listErr = state.ListPoliciesByPathPattern(pathPattern)
	default:
		policies, listErr = state.ListPolicies()
	}

	if listErr != nil {
		return net.RespondWithHTTPError(listErr, w, reqres.PolicyListResponse{})
	}

	items := make([]data.PolicyListItem, len(policies))
	for i, p := range policies {
		items[i] = data.PolicyListItem{
			ID:   p.ID,
			Name: p.Name,
		}
	}

	return net.Success(reqres.PolicyListResponse{Policies: items}.Success(), w)
}
