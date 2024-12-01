//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike/app/keeper/internal/state"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteShow handles requests to retrieve the currently cached root key.
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
func RouteShow(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeShow"
	log.AuditRequest(fName, r, audit, log.AuditRead)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RootKeyReadRequest, reqres.RootKeyReadResponse](
		requestBody, w,
		reqres.RootKeyReadResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	responseBody := net.MarshalBody(
		reqres.RootKeyReadResponse{RootKey: state.RootKey()}, w,
	)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
