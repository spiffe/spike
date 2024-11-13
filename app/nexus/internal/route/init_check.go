//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routeInitCheck handles HTTP requests to check the initialization state of
// the system. It determines whether the system has been initialized by checking
// for the presence of an admin token. This endpoint can be accessed anonymously,
// as it is used during the initial system setup process by the first user who
// will become the administrator.
//
// The function performs the following steps:
//  1. Logs the incoming request details
//  2. Reads and validates the request body
//  3. Checks for an existing admin token
//  4. Returns the appropriate initialization state in the response
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: nil if the check completes successfully, or an error if:
//   - Request body cannot be read
//   - Response body cannot be marshaled
//   - System is already initialized (returns "already initialized" error)
//
// Response Status Codes:
//   - 200 OK: Successfully checked initialization state
//
// Response Body:
//   - JSON object containing:
//   - State: Either "already_initialized" or "not_initialized"
//
// Example Response:
//
//	{
//	  "state": "not_initialized"
//	}
//
// Note: This endpoint intentionally skips JWT validation as it needs to be
// accessible during initial system setup.
func routeInitCheck(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeInitCheck", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = "init-check"

	// No need to check for valid JWT here. System initialization is done
	// anonymously by the first user (who will be the admin).
	// If the system is already initialized, this process will err out anyway.

	responseBody := net.ReadRequestBody(r, w)
	if responseBody == nil {
		return errors.New("failed to read request body")
	}

	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInitCheck", "msg", "Already initialized")

		responseBody := net.MarshalBody(reqres.CheckInitStateResponse{
			State: data.AlreadyInitialized}, w,
		)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusOK, responseBody, w)
		log.Log().Info("routeInitCheck",
			"already_initialized", true,
			"msg", "OK",
		)
		return errors.New("already initialized")
	}

	responseBody = net.MarshalBody(reqres.CheckInitStateResponse{
		State: data.NotInitialized,
	}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInitCheck", "msg", "OK")
	return nil
}
