//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/retry"
)

// ReadAdminRecoveryMetadata retrieves cached admin recovery metadata from
// the backend storage. It attempts to load the credentials using a context
// with timeout specified by DatabaseOperationTimeout. If the backend is
// unavailable or there's an error loading credentials, it returns nil and
// logs a warning, treating memory as the source of truth.
//
// Returns:
//   - *data.RecoveryMetadata: The cached admin recovery metadata if
//     successfully loaded, nil otherwise.
func ReadAdminRecoveryMetadata() *data.RecoveryMetadata {
	be := Backend()
	if be == nil {
		return nil
	}

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[data.RecoveryMetadata](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	metadata, err := typedRetrier.RetryWithBackoff(ctx, func() (data.RecoveryMetadata, error) {
		return be.LoadAdminRecoveryMetadata(ctx)
	})

	if err != nil {
		log.Log().Warn("readAdminRecoveryMetadata",
			"msg", "Failed to load admin recovery metadata after retries",
			"err", err.Error(),
		)
		return nil
	}

	return &metadata
}

// AsyncPersistAdminRecoveryMetadata asynchronously stores admin recovery
// metadata in the backend storage. It launches a goroutine to perform the
// storage operation with a timeout specified by DatabaseOperationTimeout.
// If the backend is unavailable or there's an error storing credentials,
// it logs a warning and continues execution, as memory is considered the
// source of truth.
//
// Parameters:
//   - credentials: data.RecoveryMetadata to be stored in the backend
func AsyncPersistAdminRecoveryMetadata(credentials data.RecoveryMetadata) {
	go func() {
		be := Backend()
		if be == nil {
			return // No cache available
		}

		retrier := retry.NewExponentialRetrier()

		ctx, cancel := context.WithTimeout(
			context.Background(), env.DatabaseOperationTimeout(),
		)
		defer cancel()

		err := retrier.RetryWithBackoff(ctx, func() error {
			return be.StoreAdminRecoveryMetadata(ctx, credentials)
		})

		if err != nil {
			log.Log().Warn("asyncPersistAdminRecoveryMetadata",
				"msg", "Failed to cache admin recovery metadata after retries",
				"err", err.Error(),
			)
		}
	}()
}
