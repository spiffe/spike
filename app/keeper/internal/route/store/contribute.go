//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// TODO: likely, there's lots of godoc changes too; check the PR for those.

// RouteContribute handles HTTP requests for shard contributions in the system.
// It processes incoming shard data, decodes it from Base64 encoding, and stores
// it in the system state.
//
// The function expects a Base64-encoded shard and a keeper ID in the request
// body. It performs the following operations:
//   - Reads and validates the request body
//   - Decodes the Base64-encoded shard
//   - Stores the decoded shard in the system state
//   - Logs the operation for auditing purposes
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for tracking the request for auditing purposes
//
// Returns:
//   - error: nil if successful, otherwise one of:
//   - errors.ErrReadFailure if request body cannot be read
//   - errors.ErrParseFailure if request parsing fails or shard decoding fails
//
// Example request body:
//
//	{
//	  "shard": "base64EncodedString",
//	  "keeperId": "uniqueIdentifier"
//	}
//
// The function returns a 200 OK status with an empty response body on success,
// or a 400 Bad Request status with an error message if the shard content is
// invalid.
func RouteContribute(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeContribute"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.ShardContributionRequest, reqres.ShardContributionResponse](
		requestBody, w,
		reqres.ShardContributionResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	shard := request.Shard
	id := request.KeeperId
	// Security: Zero out shard before the function exits.
	defer func() {
		for i := 0; i < len(shard); i++ {
			shard[i] = 0
		}
	}()

	// Decode shard content from Base64 encoding.
	// decodedShard, err := base64.StdEncoding.DecodeString(shard)
	//if err != nil {
	//	log.Log().Error(fName, "msg", "Failed to decode shard", "err", err.Error())
	//	http.Error(w, "Invalid shard content", http.StatusBadRequest)
	//	return errors.ErrParseFailure
	//}

	zeroed := true
	for _, c := range shard {
		if c != 0 {
			zeroed = false
			break
		}
	}

	if zeroed {
		responseBody := net.MarshalBody(reqres.ShardContributionResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)

		return errors.ErrParseFailure
	}

	state.SetShard(&shard)
	log.Log().Info(fName, "msg", "Shard stored", "id", id)

	responseBody := net.MarshalBody(reqres.ShardContributionResponse{}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)

	return nil
}
