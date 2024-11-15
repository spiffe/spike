//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routeKeep handles caching a root key received from SPIKE Nexus.
//
// This endpoint is authenticated via SPIFFE for machine-to-machine
// communication between SPIKE Keeper and SPIKE Nexus, rather than using JWT
// tokens. The function:
//  1. Logs the incoming request details
//  2. Reads and validates the request body
//  3. Extracts the root key
//  4. Caches it in the application state
//  5. Returns a success response
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
//	    "rootKey": string   // The root key to cache
//	}
//
// On success, returns an empty 200 OK response. Returns 400 Bad Request if the
// request body is invalid or missing required fields.
func routeKeep(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeKeep", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = "create"

	// Note: no JWT validation is performed here because SPIKE Keeper trusts
	// SPIKE Nexus through SPIFFE authentication. There is no human user
	// involved in this request, so no JWT is needed.

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.RootKeyCacheRequest, reqres.RootKeyCacheResponse](
		requestBody, w,
		reqres.RootKeyCacheResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	rootKey := request.RootKey
	state.SetRootKey(rootKey)

	responseBody := net.MarshalBody(reqres.RootKeyCacheResponse{}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeKeep", "msg", "OK")
	return nil
}
