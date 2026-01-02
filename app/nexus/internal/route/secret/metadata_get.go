//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike-sdk-go/journal"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// RouteGetSecretMetadata handles requests to retrieve secret metadata at a
// specific path and version.
//
// This endpoint requires the peer to have read permission for the specified
// secret path. The function retrieves secret metadata based on the provided
// path and optional version number. If no version is specified (version 0),
// the current version's metadata is returned.
//
// The function follows these steps:
//  1. Authenticates the peer via SPIFFE ID and validates read permissions
//  2. Validates and unmarshals the request body
//  3. Attempts to retrieve the secret metadata
//  4. Returns the secret metadata or an appropriate error response
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request with peer SPIFFE ID
//   - audit: *journal.AuditEntry for logging audit information
//
// Returns:
//   - *sdkErrors.SDKError: Returns nil on successful execution, or an error
//     describing what went wrong
//
// Request body format:
//
//	{
//	    "path": string,     // Path to the secret
//	    "version": int      // Optional: specific version to retrieve
//	                        // (0 = current)
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
//   - 200 OK: Secret metadata successfully retrieved
//   - 400 Bad Request: Invalid request body or path format
//   - 401 Unauthorized: Authentication or authorization failure
//   - 404 Not Found: Secret does not exist at specified path/version
//   - 500 Internal Server Error: Backend or server-side failure
//
// All operations are logged using structured logging.
func RouteGetSecretMetadata(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeGetSecretMetadata"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	request, err := net.ReadParseAndGuard[
		reqres.SecretMetadataRequest, reqres.SecretMetadataResponse,
	](
		w, r, reqres.SecretMetadataResponse{}.BadRequest(),
		guardGetSecretMetadataRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path
	version := request.Version

	rawSecret, getErr := state.GetRawSecret(path, version)
	if getErr != nil {
		return net.RespondWithHTTPError(getErr, w, reqres.SecretMetadataResponse{})
	}

	return net.Success(toSecretMetadataSuccessResponse(rawSecret), w)
}
