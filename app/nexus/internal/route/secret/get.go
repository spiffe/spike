//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteGetSecret handles requests to retrieve a secret at a specific path
// and version.
//
// This endpoint requires a valid admin JWT token for authentication. The
// function retrieves a secret based on the provided path and optional version
// number. If no version is specified, the latest version is returned.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates and unmarshals the request body
//  3. Attempts to retrieve the secret
//  4. Returns the secret data or an appropriate error response
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
//	{
//	    "path": string,     // Path to the secret
//	    "version": int      // Optional: specific version to retrieve
//	}
//
// Response format on success (200 OK):
//
//	{
//	    "data": {          // The secret data
//	        // Secret key-value pairs
//	    }
//	}
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body
//   - 404 Not Found: Secret doesn't exist at specified path/version
//
// All operations are logged using structured logging.
func RouteGetSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetSecret"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return apiErr.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretReadRequest, reqres.SecretReadResponse](
		requestBody, w,
		reqres.SecretReadResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return apiErr.ErrParseFailure
	}

	version := request.Version
	path := request.Path

	err := guardGetSecretRequest(*request, w, r)
	if err != nil {
		return err
	}

	secret, err := state.GetSecret(path, version)
	if err == nil {
		log.Log().Info(fName, "message", "Secret found")
	}
	if err != nil {
		return handleGetSecretError(err, w)
	}

	responseBody := net.MarshalBody(reqres.SecretReadResponse{
		Secret: data.Secret{Data: secret},
	}, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetSecret", "message", data.ErrSuccess)
	return nil
}
