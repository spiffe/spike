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
	"github.com/spiffe/spike-sdk-go/strings"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteShard handles HTTP requests to retrieve the stored shard from the
// system. It retrieves the shard from the system state and returns it to the
// requester.
//
// Security:
//
// This endpoint validates that the requesting peer is SPIKE Nexus using SPIFFE
// ID verification. Only SPIKE Nexus is authorized to retrieve shards during
// recovery operations. Unauthorized requests receive a 401 Unauthorized
// response.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *journal.AuditEntry for tracking the request for auditing purposes
//
// Returns:
//   - error: nil if successful, otherwise one of:
//   - errors.ErrReadFailure if request body cannot be read
//   - errors.ErrParseFailure if request parsing fails
//   - errors.ErrUnauthorized if peer SPIFFE ID validation fails
//   - errors.ErrNotFound if no shard is stored in the system
//
// Response body:
//
//	{
//	  "shard": "base64EncodedString"
//	}
//
// The function returns a 200 OK status with the encoded shard on success,
// a 404 Not Found status if no shard exists, or a 401 Unauthorized status
// if the peer is not SPIKE Nexus.
func RouteShard(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "RouteShard"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)
	_, err := net.ReadParseAndGuard[
		reqres.ShardGetRequest, reqres.ShardGetResponse,
	](
		w, r, reqres.ShardGetBadInput, guardShardGetRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	state.RLockShard()
	defer state.RUnlockShard()
	// DO NOT reset `sh` after use, as this function does NOT "own" it.
	// Treat the value as "read-only".
	sh := state.ShardNoSync()

	if err := net.FailIf(
		mem.Zeroed32(sh),
		reqres.ShardGetBadInput, w,
		http.StatusBadRequest, errors.ErrInvalidInput,
	); err != nil {
		log.Log().Error(fName, "message", data.ErrBadInput, "err", err.Error())
		return err
	}

	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.ShardGetResponse{Shard: sh}.Success(), w,
	)
	if alreadyResponded := err != nil; !alreadyResponded {
		net.Respond(http.StatusOK, responseBody, w)
	}
	// Security: Reset response body before function exits.
	defer func() {
		mem.ClearBytes(responseBody)
	}()
	log.Log().Info(
		fName,
		"message", data.ErrSuccess,
		"err", strings.MaybeError(err),
	)
	return nil
}
