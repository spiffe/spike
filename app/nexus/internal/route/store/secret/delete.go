//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spiffe/spike/internal/entity"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/pkg/spiffe"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteDeleteSecret handles HTTP DELETE requests for secret deletion
// operations. It validates the JWT token, processes the deletion request,
// and manages the secret deletion workflow.
//
// The function expects a request body containing a path and optional version
// numbers of the secrets to be deleted. If no versions are specified, an empty
// slice is used.
//
// Parameters:
//   - w: http.ResponseWriter for writing the HTTP response
//   - r: *http.Request containing the incoming HTTP request details
//   - audit: *log.AuditEntry for logging audit information about the deletion
//     operation
//
// Returns:
//   - error: Returns nil on successful execution, or an error describing what
//     went wrong
//
// The function performs the following steps:
//  1. Validates the JWT token against the admin token
//  2. Reads and parses the request body
//  3. Processes the secret deletion
//  4. Returns a JSON response
//
// Example request body:
//
//	{
//	    "path": "secret/path",
//	    "versions": [1, 2, 3]
//	}
//
// Possible errors:
//   - "invalid or missing JWT token": When JWT validation fails
//   - "failed to read request body": When request body cannot be read
//   - "failed to parse request body": When request body is invalid
//   - "failed to marshal response body": When response cannot be serialized
func RouteDeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeDeleteSecret"
	log.AuditRequest(fName, r, audit, log.AuditDelete)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return entity.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretDeleteRequest, reqres.SecretDeleteResponse](
		requestBody, w,
		reqres.SecretDeleteResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return entity.ErrParseFailure
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	spiffeId, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretDeleteResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	allowed := state.CheckAccess(
		spiffeId.String(),
		path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretDeleteResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return entity.ErrUnauthorized
	}

	err = state.DeleteSecret(path, versions)
	if err != nil {
		log.Log().Info(fName, "msg", "Failed to delete secret", "err", err)
	} else {
		log.Log().Info(fName, "msg", "Secret deleted")
	}

	responseBody := net.MarshalBody(reqres.SecretDeleteResponse{}, w)
	if responseBody == nil {
		return entity.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", reqres.ErrSuccess)
	return nil
}
