//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/spiffe"
)

// RouteUndeleteSecret handles HTTP requests to restore previously deleted
// secrets.
//
// This endpoint requires a valid admin JWT token for authentication. It accepts
// a POST request with a JSON body containing a path to the secret and
// optionally specific versions to undelete. If no versions are specified,
// an empty version list is used.
//
// The function validates the JWT, reads and unmarshals the request body,
// processes the undelete operation, and returns a 200 OK response upon success.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
//
// Request body format:
//
//	{
//	    "path": string,    // Path to the secret to undelete
//	    "versions": []int  // Optional list of specific versions to undelete
//	}
//
// Responses:
//   - 200 OK: Secret successfully undeleted
//   - 400 Bad Request: Invalid request body or parameters
//   - 401 Unauthorized: Invalid or missing JWT token
//
// The function logs its progress at various stages using structured logging.
func RouteUndeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeUndeleteSecret"
	log.AuditRequest(fName, r, audit, log.AuditUndelete)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return entity.ErrReadFailure
	}

	req := net.HandleRequest[
		reqres.SecretUndeleteRequest, reqres.SecretUndeleteResponse](
		requestBody, w,
		reqres.SecretUndeleteResponse{Err: reqres.ErrBadInput},
	)
	if req == nil {
		return entity.ErrParseFailure
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return entity.ErrUnauthorized
	}
	allowed := state.CheckAccess(
		spiffeid.String(),
		path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return entity.ErrUnauthorized
	}

	err = state.UndeleteSecret(path, versions)
	if err != nil {
		log.Log().Info(fName, "msg", "Failed to undelete secret", "err", err)
	} else {
		log.Log().Info(fName, "msg", "Secret undeleted")
	}

	responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{}, w)
	if responseBody == nil {
		return entity.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", reqres.ErrSuccess)
	return nil
}
