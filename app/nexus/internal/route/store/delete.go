//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"errors"
	"github.com/spiffe/spike/app/nexus/internal/state/base"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
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
//	    "rootKey": string   // Root key to be cached
//	}
//
// Responses:
//   - 200 OK: Root key successfully cached
//   - 400 Bad Request: Invalid request body or parameters
//
// The function logs its progress using structured logging. Unlike other routes,
// this endpoint relies on SPIFFE authentication rather than JWT validation.
func RouteDeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeDeleteSecret", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditDelete

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return errors.New("invalid or missing JWT token")
	}

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.SecretDeleteRequest, reqres.SecretDeleteResponse](
		requestBody, w,
		reqres.SecretDeleteResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	base.DeleteSecret(path, versions)
	log.Log().Info("routeDeleteSecret", "msg", "Secret deleted")

	responseBody := net.MarshalBody(reqres.SecretDeleteResponse{}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeDeleteSecret", "msg", "OK")
	return nil
}
