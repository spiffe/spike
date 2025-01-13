//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
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

// TODO: create func ReadRecoveryInfo
