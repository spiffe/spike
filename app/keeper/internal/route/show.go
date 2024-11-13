//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
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
func routeShow(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeShow", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	// Note: no JWT validation is performed here because SPIKE Keeper trusts
	// SPIKE Nexus through SPIFFE authentication. There is no human user
	// involved in this request, so no JWT is needed.

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.RootKeyReadRequest, reqres.RootKeyReadResponse](
		requestBody, w,
		reqres.RootKeyReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	rootKey := state.RootKey()

	responseBody := net.MarshalBody(
		reqres.RootKeyReadResponse{RootKey: rootKey}, w,
	)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeShow", "msg", "OK")
}
