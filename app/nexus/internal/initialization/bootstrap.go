//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"os"
	"path"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

func bootstrap(source *workloadapi.X509Source) {
	const fName = "bootstrap"

	if source == nil {
		// If source is nil, nobody is going to recreate the source,
		// it's better to log and crash.
		log.FatalLn(fName + ": source is nil. this should not happen.")
	}

	// The tombstone file is a fast path to validate SPIKE Nexus bootstrap
	// completion. However, it's not the ultimate criterion. If we cannot
	// find a tombstone file, then we'll query existing SPIKE Keeper instances
	// for shard information until we get enough shards to reconstruct
	// the root key. If that,too fails, then a human operator will need to
	// manually re-key SPIKE Nexus.
	tombstone := path.Join(
		config.SpikeNexusDataFolder(), config.SpikeNexusTombstoneFile,
	)

	_, err := os.Stat(tombstone)

	nexusAlreadyBootstrapped := err == nil
	if nexusAlreadyBootstrapped {
		log.Log().Info(fName,
			"msg", "Tombstone file exists, SPIKE Nexus is bootstrapped",
		)

		recovery.RecoverBackingStoreUsingKeeperShards(source)
		return
	}

	bootstrapStatusCheckFailed := !os.IsNotExist(err)
	if bootstrapStatusCheckFailed {
		// This should not typically happen.

		log.Log().Warn(fName,
			"msg", "Failed to check tombstone file. Will try keeper recovery",
			"err", err,
		)

		recovery.RecoverBackingStoreUsingKeeperShards(source)
		return
	}

	// If the flow reaches here, we assume SPIKE Nexus has not bootstrapped
	// and it's day zero. Let's bootstrap SPIKE Nexus with a fresh root key:
	recovery.BootstrapBackingStoreWithNewRootKey(source)
}
