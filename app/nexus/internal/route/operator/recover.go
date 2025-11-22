//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteRecover handles HTTP requests for recovering pilot recovery shards.
//
// This function processes HTTP requests to retrieve recovery shards needed for
// a recovery operation. It reads and validates the request, retrieves the first
// two recovery shards from the pilot recovery system, and returns them in the
// response.
//
// Parameters:
//   - w http.ResponseWriter: The HTTP response writer to write the response to.
//   - r *http.Request: The incoming HTTP request.
//   - audit *journal.AuditEntry: An audit entry for logging the request.
//
// Returns:
//   - error: An error if one occurs during processing, nil otherwise.
//
// The function will return various errors in the following cases:
//   - errors.ErrReadFailure: If the request body cannot be read.
//   - errors.ErrParseFailure: If the request body cannot be parsed.
//   - errors.ErrNotFound: If fewer than 2 recovery shards are available.
//   - Any error returned by guardRecoverRequest: For request validation
//     failures.
//
// On success, the function responds with HTTP 200 OK and the first two recovery
// shards in the response body.
func RouteRecover(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeRecover"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	_, err := net.ReadParseAndGuard[
		reqres.RecoverRequest, reqres.RecoverResponse](
		w, r, reqres.RecoverBadInput, guardRecoverRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	shards := recovery.NewPilotRecoveryShards()

	// Security: reset shards before the function exits.
	defer func() {
		for i := range shards {
			mem.ClearRawBytes(shards[i])
		}
	}()

	if len(shards) < env.ShamirThresholdVal() {
		log.Log().Error(fName, "message", sdkErrors.ErrCodeShamirNotEnoughShards)
		return sdkErrors.ErrInvalidInput
	}

	// Track seen indices to check for duplicates
	seenIndices := make(map[int]bool)

	for idx, shard := range shards {
		if seenIndices[idx] {
			log.Log().Error(fName, "message", sdkErrors.ErrCodeShamirDuplicateIndex)
			// Duplicate index.
			return sdkErrors.ErrInvalidInput
		}

		// We cannot check for duplicate values, because although it's
		// astronomically unlikely, there is still a possibility of two
		// different indices having the same shard value.

		seenIndices[idx] = true

		// Check for nil pointers
		if shard == nil {
			log.Log().Error(fName, "message", sdkErrors.ErrCodeShamirNilShard)
			return sdkErrors.ErrInvalidInput
		}

		// Check for empty shards (all zeros)
		zeroed := true
		for _, b := range *shard {
			if b != 0 {
				zeroed = false
				break
			}
		}
		if zeroed {
			log.Log().Error(fName, "message", sdkErrors.ErrCodeShamirEmptyShard)
			return sdkErrors.ErrInvalidInput
		}

		// Verify shard index is within valid range:
		if idx < 1 || idx > env.ShamirSharesVal() {
			log.Log().Error(fName, "message", sdkErrors.ErrCodeShamirInvalidIndex)
			return sdkErrors.ErrInvalidInput
		}
	}

	responseBody := net.SuccessWithResponseBody(
		reqres.RecoverResponse{Shards: shards}.Success(), w, fName,
	)
	defer func() {
		mem.ClearBytes(responseBody)
	}()
	return nil
}
