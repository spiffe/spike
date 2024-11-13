//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routeUndeleteSecret handles HTTP requests to restore previously deleted
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
func routeUndeleteSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeUndeleteSecret",
		"method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	req := net.HandleRequest[
		reqres.SecretUndeleteRequest, reqres.SecretUndeleteResponse](
		requestBody, w,
		reqres.SecretUndeleteResponse{Err: reqres.ErrBadInput},
	)
	if req == nil {
		return
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.UndeleteSecret(path, versions)
	log.Log().Info("routeUndeleteSecret", "msg", "Secret undeleted")

	responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeUndeleteSecret", "msg", "OK")
}
