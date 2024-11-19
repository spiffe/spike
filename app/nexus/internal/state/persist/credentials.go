//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/internal/log"
)

// ReadAdminCredentials retrieves cached admin credentials from the backend
// storage. It attempts to load the credentials using a context with timeout
// specified by DatabaseOperationTimeout. If the backend is unavailable or
// there's an error loading credentials, it returns nil and logs a warning,
// treating memory as the source of truth.
//
// Returns:
//   - *data.Credentials: The cached admin credentials if successfully loaded,
//     nil otherwise.
func ReadAdminCredentials() *data.Credentials {
	be := Backend()
	if be == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedCreds, err := be.LoadAdminCredentials(ctx)
	if err != nil {
		// Log error but continue - memory is source of truth
		log.Log().Warn("readAdminCredentials",
			"msg", "Failed to load admin credentials from cache",
			"err", err.Error(),
		)
		return nil
	}

	return &cachedCreds
}

// AsyncPersistAdminCredentials asynchronously stores admin credentials in the
// backend storage. It launches a goroutine to perform the storage operation
// with a timeout specified by DatabaseOperationTimeout. If the backend is
// unavailable or there's an error storing credentials, it logs a warning and
// continues execution, as memory is considered the source of truth.
//
// Parameters:
//   - credentials: data.Credentials to be stored in the backend
func AsyncPersistAdminCredentials(credentials data.Credentials) {
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

		if err := be.StoreAdminCredentials(ctx, credentials); err != nil {
			// Log error but continue - memory is source of truth
			log.Log().Warn("asyncPersistAdminCredentials",
				"msg", "Failed to cache admin credentials",
				"err", err.Error(),
			)
		}
	}()
}
