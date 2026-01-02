//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike-sdk-go/journal"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
)

var (
	shards      []recovery.ShamirShard
	shardsMutex sync.RWMutex
)

// RouteRestore handles HTTP requests for restoring SPIKE Nexus using recovery
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
//   - audit *journal.AuditEntry: An audit entry for logging the request.
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
// The function responds with HTTP 200 OK in all successful cases:
//   - Shard successfully added to the collection
//   - Restoration already complete (additional shards acknowledged but ignored)
//   - Duplicate shard received (acknowledged but ignored, status shows
//     the remaining shards needed)
//
// When the last required shard is added, the function automatically triggers
// the restoration process using RestoreBackingStoreFromPilotShards.
func RouteRestore(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeRestore"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	if env.BackendStoreTypeVal() == env.Memory {
		log.Info(fName, "message", "skipping restoration: in-memory mode")
		return nil
	}

	request, err := net.ReadParseAndGuard[
		reqres.RestoreRequest, reqres.RestoreResponse](
		w, r, reqres.RestoreResponse{}.BadRequest(), guardRestoreRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	shardsMutex.Lock()
	defer shardsMutex.Unlock()

	// Check if we already have enough shards
	currentShardCount := len(shards)

	threshold := env.ShamirThresholdVal()
	restored := currentShardCount >= threshold

	if restored {
		// Already restored; acknowledge and ignore additional shards.
		return net.Success(
			reqres.RestoreResponse{
				RestorationStatus: data.RestorationStatus{
					ShardsCollected: currentShardCount,
					ShardsRemaining: 0,
					Restored:        restored,
				},
			}.Success(), w,
		)
	}

	for _, shard := range shards {
		if int(shard.ID) != request.ID {
			continue
		}

		// Duplicate shard; acknowledge and ignore.
		return net.Success(
			reqres.RestoreResponse{
				RestorationStatus: data.RestorationStatus{
					ShardsCollected: currentShardCount,
					ShardsRemaining: threshold - currentShardCount,
					Restored:        restored,
				},
			}.Success(), w,
		)
	}

	shards = append(shards, recovery.ShamirShard{
		ID:    uint64(request.ID),
		Value: request.Shard,
	})

	currentShardCount = len(shards)

	// Note: We cannot clear request.Shard because it's a pointer type,
	// and we need it later in the "restore" operation.
	// RouteRestore cleans this up when it is no longer necessary.

	// Trigger restoration if we have collected all shards
	restored = currentShardCount >= threshold
	if restored {
		recovery.RestoreBackingStoreFromPilotShards(shards)
		// Security: Zero out all shards since we have finished restoration:
		for i := range shards {
			mem.ClearRawBytes(shards[i].Value)
			shards[i].ID = 0
		}
	}

	return net.Success(
		reqres.RestoreResponse{
			RestorationStatus: data.RestorationStatus{
				ShardsCollected: currentShardCount,
				ShardsRemaining: threshold - currentShardCount,
				Restored:        restored,
			},
		}.Success(), w,
	)
}
