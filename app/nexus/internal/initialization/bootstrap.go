//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/config"
)

// bootstrap initializes the SPIKE Nexus by either recovering from the existing
// state or creating a new root key. The function determines the appropriate
// initialization // strategy based on system state indicators.
//
// Parameters:
//   - source *workloadapi.X509Source: Authentication source for SPIFFE workload
//     API
//
// The function follows this decision flow:
//  1. Check for a tombstone file as an indicator of previous bootstrapping
//  2. If tombstone exists, perform recovery using keeper shards
//  3. If the tombstone check fails for reasons other than non-existence,
//     attempt recovery
//  4. If no tombstone exists, assume first-time initialization and create a
//     new root key
//
// The function will:
//   - Log and terminate if a nil source is provided
//   - Recover existing state if SPIKE Nexus was previously bootstrapped
//   - Create a new root key if no previous state is detected
//   - Hydrate in-memory data stores after recovery or initialization
//
// Errors:
//   - Fatal error if the source parameter is nil
//   - Warning on tombstone file check errors, with fallback to recovery
func bootstrap(source *workloadapi.X509Source) {
	const fName = "bootstrap"

	if source == nil {
		// If `source` is nil, nobody is going to recreate the source,
		// it's better to log and crash.
		log.FatalLn(fName + ": source is nil. this should not happen.")
	}

	sqlBackend := env.BackendStoreType() == env.Sqlite

	// The tombstone file is a fast heuristic to validate SPIKE Nexus bootstrap
	// completion. However, it's not the ultimate criterion. If we cannot
	// find a tombstone file, then we'll query existing SPIKE Keeper instances
	// for shard information until we get enough shards to reconstruct
	// the root key. If that, too, fails, then a human operator will need to
	// manually re-key SPIKE Nexus.
	tombstone := config.SpikeNexusTombstonePath()

	_, err := os.Stat(tombstone)

	nexusAlreadyBootstrapped := err == nil
	if nexusAlreadyBootstrapped {
		log.Log().Info(fName,
			"message", "Tombstone file exists, "+
				"SPIKE Nexus is bootstrapped. Will try keeper recovery",
		)

		recovery.RecoverBackingStoreUsingKeeperShards(source)
		if sqlBackend {
			recovery.HydrateMemoryFromBackingStore()
		}
		return
	}

	bootstrapStatusCheckFailed := !os.IsNotExist(err)
	if bootstrapStatusCheckFailed {
		// This should not typically happen.

		log.Log().Warn(fName,
			"message", "Failed to check tombstone file. Will try keeper recovery",
			"err", err,
		)

		recovery.RecoverBackingStoreUsingKeeperShards(source)
		if sqlBackend {
			recovery.HydrateMemoryFromBackingStore()
		}
		return
	}

	// If the flow reaches here, we assume SPIKE Nexus has not bootstrapped
	// and it's day zero. Let's bootstrap SPIKE Nexus with a fresh root key:
	recovery.BootstrapBackingStoreWithNewRootKey(source)
}
