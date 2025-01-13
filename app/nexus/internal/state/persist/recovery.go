//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/retry"
	"github.com/spiffe/spike/pkg/store"
)

func AsyncPersistRecoveryInfo(meta store.KeyRecoveryData) {
	be := Backend()

	go func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			env.DatabaseOperationTimeout(),
		)
		defer cancel()

		if err := be.StoreKeyRecoveryInfo(ctx, meta); err != nil {
			log.Log().Warn("asyncPersistRecoveryInfo",
				"msg", "Failed to cache recovery info",
				"err", err.Error())
		}
	}()
}

func ReadRecoveryInfo() *store.KeyRecoveryData {
	be := Backend()
	if be == nil {
		return nil
	}

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[*store.KeyRecoveryData](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedRecoveryInfo, err := typedRetrier.RetryWithBackoff(ctx, func() (*store.KeyRecoveryData, error) {
		return be.LoadKeyRecoveryInfo(ctx)
	})
	if err != nil {
		log.Log().Warn("readRecoveryInfo",
			"msg", "Failed to load recovery info from cache after retries",
			"err", err.Error())
		return nil
	}

	if cachedRecoveryInfo != nil {
		return cachedRecoveryInfo
	}

	return nil
}
