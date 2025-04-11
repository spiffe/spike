//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/log"
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
//   - audit *log.AuditEntry: An audit entry for logging the request.
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
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	// TODO: some of these logs are useful. add them as log.Info() or something.
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>> IN RECOVER ROUTER")

	const fName = "routeRecover"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> requestBody is nil")
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RecoverRequest, reqres.RecoverResponse](
		requestBody, w,
		reqres.RecoverResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> request is nil")
		return errors.ErrParseFailure
	}

	err := guardRecoverRequest(*request, w, r)
	if err != nil {
		return err
	}

	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> before newPilotRecoveryShards")
	shards := recovery.NewPilotRecoveryShards()
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> after newPilotRecoveryShards: len(shards) = ", len(shards))

	// Security: reset shards before function exits.
	defer func() {
		for i := range shards {
			mem.ClearRawBytes(shards[i])
		}
	}()

	if len(shards) < env.ShamirThreshold() {
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> len(shards) < env.ShamirThreshold()", len(shards), env.ShamirThreshold())
		return errors.ErrNotFound
	}

	// Track seen indices to check for duplicates
	seenIndices := make(map[int]bool)

	for idx, shard := range shards {
		if seenIndices[idx] {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> invalid input 0001")

			// Duplicate index.
			return errors.ErrInvalidInput
		}

		// We cannot check for duplicate values, because although it's
		// astronomically unlikely, there is still a possibility of two
		// different indices having the same shard value.

		seenIndices[idx] = true

		// Check for nil pointers
		if shard == nil {
			return errors.ErrInvalidInput
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
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> invalid input 0002")

			return errors.ErrInvalidInput
		}

		// Verify shard index is within valid range:
		if idx < 1 || idx > env.ShamirShares() {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> invalid input 0003 ", idx, env.ShamirShares())

			return errors.ErrInvalidInput
		}
	}

	responseBody := net.MarshalBody(reqres.RecoverResponse{
		Shards: shards,
	}, w)
	// Security: Clean up response body before exit.
	defer func() {
		mem.ClearBytes(responseBody)
	}()

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
