//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/base64"
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
	expectedShardCount = 2
	decodedShardSize   = 32 // bytes
)

var (
	shards      []string
	shardsMutex sync.RWMutex
)

// validateShard checks if the shard is valid and not duplicate
func validateShard(shard string) error {
	// Check if shard is already stored
	shardsMutex.RLock()
	for _, existingShard := range shards {
		if existingShard == shard {
			shardsMutex.RUnlock()
			return errors.ErrInvalidInput
		}
	}
	shardsMutex.RUnlock()

	// Validate shard length
	decodedShard, err := base64.StdEncoding.DecodeString(shard)
	if err != nil {
		return errors.ErrInvalidInput
	}
	if len(decodedShard) != decodedShardSize {
		return errors.ErrInvalidInput
	}

	return nil
}

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

	if currentShardCount >= expectedShardCount {
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
				ShardsRemaining: expectedShardCount - currentShardCount,
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
	if currentShardCount == expectedShardCount {
		recovery.RestoreBackingStoreUsingPilotShards(shards)
	}

	responseBody := net.MarshalBody(reqres.RestoreResponse{
		RestorationStatus: data.RestorationStatus{
			ShardsCollected: currentShardCount,
			ShardsRemaining: expectedShardCount - currentShardCount,
			Restored:        currentShardCount == expectedShardCount,
		},
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
