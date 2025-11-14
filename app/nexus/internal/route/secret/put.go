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

// RoutePutSecret handles HTTP requests to create or update secrets at a
// specified path.
//
// This endpoint requires a valid admin JWT token for authentication. It accepts
// a PUT request with a JSON body containing the secret path and values to
// store. The function performs an upsert operation, creating a new secret if it
// doesn't exist or updating an existing one.
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
//	    "path": string,          // Path where the secret should be stored
//	    "values": map[string]any // Key-value pairs representing the secret data
//	}
//
// Responses:
//   - 200 OK: Secret successfully created or updated
//   - 400 Bad Request: Invalid request body or parameters
//   - 401 Unauthorized: Invalid or missing JWT token
//
// The function logs its progress at various stages using structured logging.
func RoutePutSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routePutSecret"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)
	request, err := net.ReadParseAndGuard[
		reqres.SecretPutRequest,
		reqres.SecretPutResponse](
		w, r,
		reqres.SecretPutResponse{Err: data.ErrBadInput},
		guardSecretPutRequest,
		fName,
	)
	alreadyResponded := err != nil
	if alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	values := request.Values
	path := request.Path

	err = state.UpsertSecret(path, values)
	if err != nil {
		return err
	}

	log.Log().Info(fName, "message", "Secret upserted")

	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.SecretPutResponse{}, w,
	)
	if err == nil {
		net.Respond(http.StatusOK, responseBody, w)
	}

	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
