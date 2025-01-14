//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"os"
	"path"
)

func Tick(source *workloadapi.X509Source) {
	// Talk to all SPIKE Keeper endpoints and send their shards and get
	// acknowledgement that they received the shard.

	if source == nil {
		// If source is nil, nobody is going to recreate the source,
		// it's better to log and crash.
		log.FatalLn("Tick: source is nil. this should not happen.")
	}

	// The tombstone file is a fast path to validate SPIKE Nexus bootstrap
	// completion. However, it's not the ultimate criterion. If we cannot
	// find a tombstone file, then we'll query existing SPIKE Keeper instances
	// for shard information until we get enough shards to reconstruct
	// the root key.
	//
	// If the keepers have crashed too, then a human operator will have to
	// manually update the Keeper instances using the "break-the-glass"
	// emergency recovery procedure as outlined in https://spike.ist/
	tombstone := path.Join(config.SpikeNexusDataFolder(), "bootstrap.tombstone")

	_, err := os.Stat(tombstone)

	nexusAlreadyBootstrapped := err == nil
	if nexusAlreadyBootstrapped {
		log.Log().Info("tick",
			"msg", "Tombstone file exists, SPIKE Nexus is bootstrapped",
		)

		recoverUsingKeeperShards(source)
	}

	// TODO: if you stop nexus, delete the tombstone file, and restart nexus,
	// it will reset its root key and update the keepers to store the new
	// root key. This is not an attack vector, because an adversary who can
	// delete the tombstone file, can also delete the backing store. In either
	// case, for production systems, the backing store needs to be backed up
	// and the root key needs to be backed up in a secure place too.
	// ^ add these to the documentation.

	bootstrapStatusCheckFailed := !os.IsNotExist(err)
	if bootstrapStatusCheckFailed {
		log.Log().Warn("tick", "msg", "Failed to check tombstone file. Will try keeper recovery", "err", err)

		recoverUsingKeeperShards(source)
		return
	}

	// TODO: if at least one Keeper returns a shard, then the system is
	// bootstrapped; do not proceed with a re-bootstrap as it will cause
	// data loss. Instead stay in  `recoverUsingKeeperShards` loop.
	// add it here as an additional check.

	// Below, SPIKE Nexus is assumed to not have bootstrapped.
	// Let's bootstrap it.

	bootstrapBackingStore(source)
}
