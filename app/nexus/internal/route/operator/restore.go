//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike/app/nexus/internal/env"
	"net/http"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

const (
	decodedShardSize = 32 // bytes
)

var (
	shards      []string
	shardsMutex sync.RWMutex
)

// RouteRestore handles HTTP requests for restoring a system using recovery
// shards.
//
// This function processes requests to contribute a recovery shard to the
// restoration process. It validates the incoming shard, adds it to the
// collection, and triggers the full restoration once all expected shards have
// been collected.
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
//   - errors.ErrMarshalFailure: If the response body cannot be marshaled.
//   - Any error returned by guardRestoreRequest: For request validation
//     failures.
//
// The function responds with:
//   - HTTP 400 Bad Request: If all required shards have already been collected
//     or if the provided shard is invalid.
//   - HTTP 200 OK: If the shard is successfully added, including status
//     information about the restoration progress.
//
// When the last required shard is added, the function automatically triggers
// the restoration process using RestoreBackingStoreUsingPilotShards.
func RouteRestore(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeRestore"

	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RestoreRequest, reqres.RestoreResponse](
		requestBody, w,
		reqres.RestoreResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	err := guardRestoreRequest(*request, w, r)
	if err != nil {
		return err
	}

	// Check if we already have enough shards
	shardsMutex.RLock()
	currentShardCount := len(shards)
	shardsMutex.RUnlock()

	if currentShardCount >= env.ShamirThreshold() {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			RestorationStatus: data.RestorationStatus{
				ShardsCollected: currentShardCount,
				ShardsRemaining: 0,
				Restored:        true,
			},
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}

	// Validate the new shard
	if err := validateShard(request.Shard); err != nil {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			RestorationStatus: data.RestorationStatus{
				ShardsCollected: currentShardCount,
				ShardsRemaining: env.ShamirThreshold() - currentShardCount,
				Restored:        false,
			},
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}

	// Add the new shard
	shardsMutex.Lock()
	shards = append(shards, request.Shard)
	currentShardCount = len(shards)
	shardsMutex.Unlock()

	// Trigger restoration if we have collected all shards
	if currentShardCount == env.ShamirThreshold() {
		recovery.RestoreBackingStoreUsingPilotShards(shards)
	}

	responseBody := net.MarshalBody(reqres.RestoreResponse{
		RestorationStatus: data.RestorationStatus{
			ShardsCollected: currentShardCount,
			ShardsRemaining: env.ShamirThreshold() - currentShardCount,
			Restored:        currentShardCount == env.ShamirThreshold(),
		},
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
