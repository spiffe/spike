//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/retry"
)

// TODO: document public methods including this.
// But before that make all asnyc persist methods sync.

func AsyncPersistPolicy(policy data.Policy) {
	be := Backend()

	if policy.Id != "" {
		go func() {
			if be == nil {
				return // No cache available
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				env.DatabaseOperationTimeout(),
			)
			defer cancel()

			if err := be.StorePolicy(ctx, policy); err != nil {
				// Log error but continue - memory is source of truth
				log.Log().Warn("asyncPersistPolicy",
					"msg", "Failed to cache policy",
					"id", policy.Id,
					"err", err.Error(),
				)
			}
		}()
	}
}

func AsyncDeletePolicy(id string) {
	be := Backend()

	if id != "" {
		go func() {
			if be == nil {
				return // No cache available
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				env.DatabaseOperationTimeout(),
			)
			defer cancel()

			if err := be.DeletePolicy(ctx, id); err != nil {
				// Log error but continue - memory is source of truth
				log.Log().Warn("asyncDeletePolicy",
					"msg", "Failed to delete policy from cache",
					"id", id,
					"err", err.Error(),
				)
			}
		}()
	}
}

func ReadPolicy(id string) *data.Policy {
	be := Backend()
	if be == nil {
		return nil
	}

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[*data.Policy](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedPolicy, err := typedRetrier.RetryWithBackoff(ctx, func() (*data.Policy, error) {
		return be.LoadPolicy(ctx, id)
	})
	if err != nil {
		log.Log().Warn("readPolicy",
			"msg", "Failed to load policy from cache after retries",
			"id", id,
			"err", err.Error(),
		)
		return nil
	}

	return cachedPolicy
}
