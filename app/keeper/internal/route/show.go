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

// routeShow handles requests to retrieve the currently cached root key.
//
// This endpoint is authenticated via SPIFFE for machine-to-machine
// communication between SPIKE Keeper and SPIKE Nexus, rather than using JWT
// tokens. The function:
//  1. Logs the incoming request details
//  2. Validates the request body format
//  3. Retrieves the current root key from application state
//  4. Returns the root key in the response
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
//	{} // Empty request body expected
//
// Response format:
//
//	{
//	    "rootKey": string   // The currently cached root key
//	}
//
// Returns 200 OK on success with the root key, or 400 Bad Request if the
// request body is malformed. Authentication is handled via SPIFFE rather
// than JWT as this is an internal system endpoint.
func routeShow(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeShow", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = "read"

	// Note: no JWT validation is performed here because SPIKE Keeper trusts
	// SPIKE Nexus through SPIFFE authentication. There is no human user
	// involved in this request, so no JWT is needed.

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.RootKeyReadRequest, reqres.RootKeyReadResponse](
		requestBody, w,
		reqres.RootKeyReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	responseBody := net.MarshalBody(
		reqres.RootKeyReadResponse{RootKey: state.RootKey()}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeShow", "msg", "OK")
	return nil
}
