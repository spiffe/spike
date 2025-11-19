//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	stdErrs "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// TODO: either return an erro, or log, but not both: verify this across the codebase
// Only exception is fatal exits.
// When you are logging an error, you are handling it.
// Errors shall ideally be handled once.

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
//   - `spiffe_id_pattern` and `path_pattern` used together
//   - Failed to marshal response body
func RouteListPolicies(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	fName := "RouteListPolicies"

	journal.AuditRequest(fName, r, audit, journal.AuditList)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyListRequest, reqres.PolicyListResponse](
		w, r, reqres.PolicyListBadInput, guardListPolicyRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	var policies []data.Policy

	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern

	// Note that Go's default switch behavior will not fall through.
	switch {
	case SPIFFEIDPattern != "":
		policies, err = state.ListPoliciesBySPIFFEIDPattern(SPIFFEIDPattern)
		if err != nil {
			failErr := sdkErrors.ErrStoreQueryFailure.Wrap(err)
			return net.Fail(
				reqres.PolicyListInternal, w,
				http.StatusInternalServerError, failErr, fName,
			)
		}
	case pathPattern != "":
		policies, err = state.ListPoliciesByPathPattern(pathPattern)
		if err != nil {
			failErr := stdErrs.Join(sdkErrors.ErrQueryFailure, err)
			return net.Fail(
				reqres.PolicyListInternal, w,
				http.StatusInternalServerError, failErr, fName,
			)
		}
	default:
		policies, err = state.ListPolicies()
		if err != nil {
			failErr := stdErrs.Join(sdkErrors.ErrQueryFailure, err)
			return net.Fail(
				reqres.PolicyListInternal, w,
				http.StatusInternalServerError, failErr, fName,
			)
		}
	}

	net.Success(reqres.PolicyListResponse{Policies: policies}.Success(), w, fName)
	return nil
}
