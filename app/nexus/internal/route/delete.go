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

// routeKeep handles HTTP requests to cache a root key from SPIKE Nexus.
//
// This endpoint is specifically designed for internal system communication
// between SPIKE Keeper and SPIKE Nexus. Authentication is handled via SPIFFE
// rather than JWT tokens, as this is a machine-to-machine interaction without
// human user involvement.
//
// Request body format:
//
//	{
//	    "rootKey": string   // Root key to be cached
//	}
//
// Responses:
//   - 200 OK: Root key successfully cached
//   - 400 Bad Request: Invalid request body or parameters
//
// The function logs its progress using structured logging. Unlike other routes,
// this endpoint relies on SPIFFE authentication rather than JWT validation.
func routeDeleteSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeDeleteSecret", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.SecretDeleteRequest, reqres.SecretDeleteResponse](
		requestBody, w,
		reqres.SecretDeleteResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.DeleteSecret(path, versions)
	log.Log().Info("routeDeleteSecret", "msg", "Secret deleted")

	responseBody := net.MarshalBody(reqres.SecretDeleteResponse{}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeDeleteSecret", "msg", "OK")
}
