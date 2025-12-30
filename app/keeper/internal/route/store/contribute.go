//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/journal"
)

// RouteContribute handles HTTP requests for the shard contributions in the
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
//   - audit: *journal.AuditEntry for tracking the request for auditing
//     purposes
//
// Returns:
//   - *sdkErrors.SDKError: nil if successful, otherwise one of:
//   - ErrDataReadFailure: If request body cannot be read
//   - ErrDataParseFailure: If request parsing fails
//   - ErrUnauthorized: If peer SPIFFE ID validation fails
//   - ErrShamirNilShard: If shard is nil
//   - ErrShamirEmptyShard: If shard is all zeros
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
) *sdkErrors.SDKError {
	const fName = "RouteContribute"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	request, err := net.ReadParseAndGuard[
		reqres.ShardPutRequest, reqres.ShardPutResponse,
	](
		w, r, reqres.ShardPutResponse{}.BadRequest(), guardShardPutRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	if request.Shard == nil {
		failErr := net.Fail(
			reqres.ShardPutResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrShamirNilShard
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
		failErr := net.Fail(
			reqres.ShardPutResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrDataInvalidInput.Wrap(failErr)
		}
		return sdkErrors.ErrShamirEmptyShard
	}

	// `state.SetShard` copies the shard. We can safely reset this one at [1].
	state.SetShard(request.Shard)

	return net.Success(reqres.ShardPutResponse{}.Success(), w)
}
