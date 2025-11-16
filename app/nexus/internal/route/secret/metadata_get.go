//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteGetSecretMetadata handles requests to retrieve a secret metadata at a
// specific path and version.
//
// This endpoint requires a valid admin JWT token for authentication. The
// function retrieves a secret based on the provided path and optional version
// number. If no version is specified, the latest version is returned.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates and unmarshal the request body
//  3. Attempts to retrieve the secret metadata
//  4. Returns the secret metadata or an appropriate error response
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
// "versions": {          // map[int]SecretMetadataVersionResponse
//
//	"version": {          // SecretMetadataVersionResponse object
//	  "createdTime": "",  // time.Time
//	  "version": 0,       // int
//	  "deletedTime": null // *time.Time (pointer, can be null)
//	 }
//	},
//
// "metadata": {          // SecretRawMetadataResponse object
//
//	 "currentVersion": 0, // int
//	 "oldestVersion": 0,  // int
//	 "createdTime": "",   // time.Time
//	 "updatedTime": "",   // time.Time
//	 "maxVersions": 0     // int
//	},
//
// "err": null            // ErrorCode
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body
//   - 404 Not Found: Secret doesn't exist at specified path/version
//
// All operations are logged using structured logging.
func RouteGetSecretMetadata(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetSecretMetadata"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	request, err := net.ReadParseAndGuard[
		reqres.SecretMetadataRequest, reqres.SecretMetadataResponse,
	](
		w, r, reqres.SecretMetadataBadInput, guardGetSecretMetadataRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	path := request.Path
	version := request.Version

	rawSecret, err := state.GetRawSecret(path, version)
	if err != nil {
		return handleGetSecretMetadataError(err, w)
	}

	response := toSecretMetadataResponse(rawSecret)
	net.Success(response.Success(), w, fName)
	return nil
}
