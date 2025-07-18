//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/keeper/internal/state"
	journal "github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteShard handles HTTP requests to retrieve the stored shard from the
// system. It retrieves the shard from the system state, encodes it in Base64,
// and returns it to the requester.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for tracking the request for auditing purposes
//
// Returns:
//   - error: nil if successful, otherwise one of:
//   - errors.ErrReadFailure if request body cannot be read
//   - errors.ErrParseFailure if request parsing fails
//   - errors.ErrNotFound if no shard is stored in the system
//
// Response body:
//
//	{
//	  "shard": "base64EncodedString"
//	}
//
// The function returns a 200 OK status with the encoded shard on success,
// or a 404 Not Found status if no shard exists in the system.
func RouteShard(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeShard"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.ShardRequest, reqres.ShardResponse](
		requestBody, w,
		reqres.ShardResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	state.RLockShard()
	defer state.RUnlockShard()
	// DO NOT reset `sh` after use, as this function does NOT own it.
	// Treat the value as "read-only".
	sh := state.ShardNoSync()

	if mem.Zeroed32(sh) {
		log.Log().Error(fName, "msg", "No shard found")

		responseBody := net.MarshalBody(reqres.ShardResponse{
			Err: data.ErrNotFound,
		}, w)
		net.Respond(http.StatusNotFound, responseBody, w)

		return errors.ErrNotFound
	}

	responseBody := net.MarshalBody(reqres.ShardResponse{
		Shard: sh,
	}, w)
	// Security: Reset response body before function exits.
	defer func() {
		mem.ClearBytes(responseBody)
	}()

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)

	return nil
}
