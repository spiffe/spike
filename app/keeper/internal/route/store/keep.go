//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/entity"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteKeep handles caching a root key received from SPIKE Nexus.
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
func RouteKeep(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeKeep"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return entity.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RootKeyCacheRequest, reqres.RootKeyCacheResponse](
		requestBody, w,
		reqres.RootKeyCacheResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return entity.ErrParseFailure
	}

	rootKey := request.RootKey
	if rootKey == "" {
		return entity.ErrMissingRootKey
	}

	state.SetRootKey(rootKey)

	responseBody := net.MarshalBody(reqres.RootKeyCacheResponse{}, w)
	if responseBody == nil {
		return entity.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", reqres.ErrSuccess)
	return nil
}
