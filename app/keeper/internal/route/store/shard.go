//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike-sdk-go/journal"
	"github.com/spiffe/spike/app/keeper/internal/state"
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
) *sdkErrors.SDKError {
	const fName = "RouteShard"

	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	_, err := net.ReadParseAndGuard[
		reqres.ShardGetRequest, reqres.ShardGetResponse,
	](
		w, r, reqres.ShardGetResponse{}.BadRequest(), guardShardGetRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	state.RLockShard()
	defer state.RUnlockShard()
	// DO NOT reset `sh` after use, as this function does NOT "own" it.
	// Treat the value as "read-only".
	sh := state.ShardNoSync()

	if mem.Zeroed32(sh) {
		failErr := net.Fail(
			reqres.ShardGetResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrDataInvalidInput
	}

	responseBody, respondErr := net.SuccessWithResponseBody(
		reqres.ShardGetResponse{Shard: sh}.Success(), w,
	)
	if respondErr != nil {
		respondErr.Msg = "failed to return success response"
		log.WarnErr(fName, *respondErr)
	}
	// Security: Reset response body before function exits.
	defer func() {
		mem.ClearBytes(responseBody)
	}()
	return nil
}
