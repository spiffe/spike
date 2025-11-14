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

// RouteContribute handles HTTP requests for shard contributions in the
// system. It processes incoming shard data and stores it in the system state.
//
// Security:
//
// This endpoint validates that the peer is either SPIKE Bootstrap or SPIKE
// Nexus using SPIFFE ID verification. SPIKE Bootstrap contributes shards
// during initial system setup, while SPIKE Nexus contributes shards during
// periodic updates. Unauthorized requests receive a 401 Unauthorized response.
//
// The function expects a shard in the request body. It performs the following
// operations:
//   - Reads and validates the request body
//   - Validates the peer SPIFFE ID
//   - Validates the shard is not nil or all zeros
//   - Stores the shard in the system state
//   - Logs the operation for auditing purposes
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
//   - errors.ErrInvalidInput if shard is nil or all zeros
//
// Example request body:
//
//	{
//	  "shard": "base64EncodedString",
//	  "keeperId": "uniqueIdentifier"
//	}
//
// The function returns a 200 OK status on success, a 401 Unauthorized status
// if the peer is not SPIKE Bootstrap or SPIKE Nexus, or a 400 Bad Request
// status if the shard content is invalid.
func RouteContribute(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "RouteContribute"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)
	request, err := net.ReadParseAndGuard[
		reqres.ShardPutRequest,
		reqres.ShardPutResponse](
		w, r,
		reqres.ShardPutResponse{Err: data.ErrBadInput},
		guardShardPutRequest,
		fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	if request.Shard == nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.ShardPutResponse{
				Err: data.ErrBadInput,
			}, w)
		if alreadyResponded := err != nil; !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		log.Log().Error(
			fName,
			"message", data.ErrBadInput,
			"err", strings.MaybeError(err),
		)
		return errors.ErrInvalidInput
	}

	// Security: Zero out shard before the function exits.
	// [1]
	defer func() {
		mem.ClearRawBytes(request.Shard)
	}()

	// Ensure the client didn't send an array of all zeros, which would
	// indicate invalid input. Since Shard is a fixed-length array in the request,
	// clients must send meaningful non-zero data.
	if mem.Zeroed32(request.Shard) {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.ShardPutResponse{
				Err: data.ErrBadInput,
			}, w)
		if alreadyResponded := err != nil; !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		log.Log().Error(
			fName,
			"message", data.ErrBadInput,
			"err", strings.MaybeError(err),
		)
		return errors.ErrInvalidInput
	}

	// `state.SetShard` copies the shard. We can safely reset this one at [1].
	state.SetShard(request.Shard)

	responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.ShardPutResponse{}, w)
	if alreadyResponded := err != nil; !alreadyResponded {
		net.Respond(http.StatusOK, responseBody, w)
	}
	log.Log().Info(
		fName,
		"message", data.ErrSuccess,
		"err", strings.MaybeError(err),
	)
	return nil
}
