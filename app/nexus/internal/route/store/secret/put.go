package secret

//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

import (
	"github.com/spiffe/spike/internal/entity"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/pkg/spiffe"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
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
//   - audit: *log.AuditEntry for logging audit information
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
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routePutSecret"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return entity.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretPutRequest, reqres.SecretPutResponse](
		requestBody, w,
		reqres.SecretPutResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return entity.ErrParseFailure
	}

	values := request.Values
	path := request.Path

	// TODO: we'll likely repeat this in a lot of places, and it can be made
	// a reusable function; maybe using generics too.
	spiffeId, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	allowed := state.CheckAccess(
		spiffeId.String(),
		path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: reqres.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return entity.ErrUnauthorized
	}

	state.UpsertSecret(path, values)
	log.Log().Info(fName, "msg", "Secret upserted")

	responseBody := net.MarshalBody(reqres.SecretPutResponse{}, w)
	if responseBody == nil {
		return entity.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", reqres.ErrSuccess)
	return nil
}
