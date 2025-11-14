//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteListPaths handles requests to retrieve all available secret paths.
//
// This endpoint requires a valid admin JWT token for authentication.
// The function returns a list of all paths where secrets are stored, regardless
// of their version or deletion status.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates the request body format
//  3. Retrieves all secret paths from the state
//  4. Returns the list of paths
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *journal.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
//
// Request body format:
//
//	{} // Empty request body expected
//
// The response format on success (200 OK):
//
//	{
//	    "keys": []string   // Array of all secret paths
//	}
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body format
//
// All operations are logged using structured logging. This endpoint only
// returns the paths to secrets and not their contents; use RouteGetSecret to
// retrieve actual secret values.
func RouteListPaths(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeListPaths"
	journal.AuditRequest(fName, r, audit, journal.AuditList)
	_, err := net.ReadParseAndGuard[
		reqres.SecretListRequest,
		reqres.SecretListResponse](
		w, r,
		reqres.SecretListResponse{Err: data.ErrBadInput},
		guardListSecretRequest,
		fName,
	)
	alreadyResponded := err != nil
	if alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	keys := state.ListKeys()
	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.SecretListResponse{Keys: keys}, w,
	)
	if err == nil {
		net.Respond(http.StatusOK, responseBody, w)
	}

	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
