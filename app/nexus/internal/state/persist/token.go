//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"sync"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/internal/log"
)

var (
	memoryBackend *memory.NoopStore
	sqliteBackend *sqlite.DataStore
	backendMu     sync.RWMutex
)

// ReadAdminSigningToken attempts to retrieve the admin token from the backend cache.
// It uses a 5-second timeout for the operation.
//
// Returns:
//   - The cached admin token as a string if successful
//   - An empty string if:
//   - The backend is not available
//   - The load operation fails
//   - No token is found in the cache
//
// Errors during the load operation are logged but not returned, as the system
// considers the in-memory state to be the source of truth.
func ReadAdminSigningToken() string {
	be := Backend()
	if be == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedToken, err := be.LoadAdminSigningToken(ctx)
	if err != nil {
		// Log error but continue - memory is source of truth
		log.Log().Warn("readAdminSigningToken",
			"msg", "Failed to load admin token from cache",
			"err", err.Error(),
		)
		return ""
	}

	return cachedToken
}

// AsyncPersistAdminToken asynchronously stores the provided admin token in the
// backend cache. The operation is performed in a background goroutine with a
// 5-second timeout.
//
// Parameters:
//   - token: The admin token to be cached
//
// The function returns immediately while the storage operation continues
// asynchronously. Any errors during the storage operation are logged but not
// returned, as the in-memory token remains the source of truth. If the backend
// is not available, the goroutine terminates without attempting storage.
func AsyncPersistAdminToken(token string) {
	go func() {
		be := Backend()
		if be == nil {
			return // No cache available
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			env.DatabaseOperationTimeout(),
		)
		defer cancel()

		if err := be.StoreAdminToken(ctx, token); err != nil {
			// Log error but continue - memory is source of truth
			log.Log().Warn("asyncPersistAdminToken",
				"msg", "Failed to cache admin token",
				"err", err.Error(),
			)
		}
	}()
}
