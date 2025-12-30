//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
)

// RoutePutSecret handles HTTP requests to create or update secrets at a
// specified path.
//
// This endpoint requires authentication via SPIFFE ID and write permission for
// the specified secret path. It accepts a PUT request with a JSON body
// containing the secret path and values to store. The function performs an
// upsert operation, creating a new secret if it doesn't exist or updating an
// existing one.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *journal.AuditEntry for logging audit information
//
// Returns:
//   - nil if the secret is successfully created or updated
//   - sdkErrors.ErrAPIPostFailed if the upsert operation fails
//   - SDK errors from request parsing or validation
//
// Request body format:
//
//	{
//	    "path": string,          // Path where the secret should be stored
//	    "values": map[string]any // Key-value pairs representing the secret data
//	}
//
// Responses:
//   - 200 OK: Secret successfully created or updated
//   - 400 Bad Request: Invalid request body or parameters
//   - 401 Unauthorized: Missing SPIFFE ID or insufficient permissions
//   - 500 Internal Server Error: Database operation failure
//
// The function logs its progress at various stages using structured logging.
func RoutePutSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RoutePutSecret"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	request, err := net.ReadParseAndGuard[
		reqres.SecretPutRequest, reqres.SecretPutResponse,
	](
		w, r, reqres.SecretPutResponse{}.BadRequest(), guardSecretPutRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	values := request.Values
	path := request.Path

	upsertErr := state.UpsertSecret(path, values)
	if upsertErr != nil {
		return net.RespondWithHTTPError(upsertErr, w, reqres.SecretPutResponse{})
	}

	return net.Success(reqres.SecretPutResponse{}.Success(), w)
}
