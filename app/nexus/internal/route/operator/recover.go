//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"

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
	const fName = "routeRecover"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RecoverRequest, reqres.RecoverResponse](
		requestBody, w,
		reqres.RecoverResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	// TODO: do I need more sanitization here?
	err := guardRecoverRequest(*request, w, r)
	if err != nil {
		return err
	}

	shards := recovery.NewPilotRecoveryShards()
	// Security: reset shards before function exits.
	defer func() {
		for i := range shards {
			for j := range shards[i][:] {
				shards[i][j] = 0
			}
		}
	}()

	if len(shards) < env.ShamirThreshold() {
		return errors.ErrNotFound
	}

	payload := make(map[int]*[32]byte)
	for i := range shards {
		payload[i] = shards[i]
	}
	// Security: Clean up the payload before exiting the function.
	defer func() {
		for i := range payload {
			payload[i] = &[32]byte{}
		}
	}()

	responseBody := net.MarshalBody(reqres.RecoverResponse{
		Shards: payload,
	}, w)
	// Security: Clean up response body before exit.
	defer func() {
		for i := range responseBody {
			responseBody[i] = 0
		}
	}()

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
