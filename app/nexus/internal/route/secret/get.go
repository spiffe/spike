//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike-sdk-go/journal"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// RouteGetSecret handles requests to retrieve a secret at a specific path
// and version.
//
// This endpoint requires the peer to have read permission for the specified
// secret path. The function retrieves a secret based on the provided path and
// optional version number. If no version is specified (version 0), the current
// version is returned.
//
// The function follows these steps:
//  1. Validates peer SPIFFE ID, authorization, and path format
//  2. Validates and unmarshals the request body
//  3. Attempts to retrieve the secret from state
//  4. Returns the secret data or an appropriate error response
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request containing the peer SPIFFE ID
//   - audit: The audit entry for logging audit information
//
// Returns:
//   - *sdkErrors.SDKError: An error if validation or retrieval fails. Returns
//     nil on success.
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
//   - 401 Unauthorized: Authentication or authorization failure
//   - 400 Bad Request: Invalid request body or path format
//   - 404 Not Found: Secret does not exist at specified path/version
//
// All operations are logged using structured logging.
func RouteGetSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeGetSecret"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	request, err := net.ReadParseAndGuard[
		reqres.SecretGetRequest, reqres.SecretGetResponse](
		w, r, reqres.SecretGetResponse{}.BadRequest(), guardGetSecretRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	version := request.Version
	path := request.Path

	secret, getErr := state.GetSecret(path, version)
	secretFound := getErr == nil

	// Extra logging to help with debugging and detecting enumeration attacks.
	if !secretFound {
		notFoundErr := sdkErrors.ErrAPINotFound.Wrap(getErr)
		notFoundErr.Msg = fmt.Sprintf(
			"secret not found at path: %s version: %d", path, version,
		)
		log.DebugErr(fName, *notFoundErr)
	}

	if !secretFound {
		return net.RespondWithHTTPError(getErr, w, reqres.SecretGetResponse{})
	}

	return net.Success(reqres.SecretGetResponse{
		Secret: data.Secret{Data: secret},
	}.Success(), w)
}
