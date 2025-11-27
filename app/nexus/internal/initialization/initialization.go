//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// Initialize initializes the SPIKE Nexus backing store based on the configured
// backend store type. The function handles three initialization modes:
//
// 1. SPIKE-Keeper-based initialization (SQLite and Lite backends):
//   - Initializes the backing store from SPIKE Keeper instances
//   - Starts a background goroutine for periodic shard synchronization
//
// 2. In-memory initialization (Memory backend):
//   - Initializes an empty in-memory backing store without root key
//   - Logs warnings about non-production use
//   - Does not use SPIKE Keepers for disaster recovery.
//
// 3. Invalid backend type:
//   - Terminates the program with a fatal error
//
// Parameters:
//   - source: An X509Source that provides X.509 certificates and private keys
//     for SPIFFE-based mTLS authentication when communicating with SPIKE
//     Keepers. Can be nil. Only used for SQLite and Lite backend types.
//     For memory backend, this parameter is ignored. For SQLite/Lite backends,
//     if source is nil, the recovery functions will log warnings and retry
//     until a valid source becomes available.
//
// Backend type configuration is determined by env.BackendStoreType().
// Valid backend types are: 'sqlite', 'lite', or 'memory'.
//
// The function will call log.FatalLn and terminate the program if an invalid
// backend store type is configured.
func Initialize(source *workloadapi.X509Source) {
	const fName = "Initialize"

	requireBackingStoreToBootstrap := env.BackendStoreTypeVal() == env.Sqlite ||
		env.BackendStoreTypeVal() == env.Lite

	if requireBackingStoreToBootstrap {
		// Initialize the backing store from SPIKE Keeper instances.
		// This is only required when the SPIKE Nexus needs bootstrapping.
		// For modes where bootstrapping is not required (such as in-memory mode),
		// SPIKE Nexus should be initialized internally.
		recovery.InitializeBackingStoreFromKeepers(source)

		// Lazy evaluation in a loop:
		// If bootstrapping is successful, start a background process to
		// periodically sync shards.
		go recovery.SendShardsPeriodically(source)

		return
	}

	devMode := env.BackendStoreTypeVal() == env.Memory

	if devMode {
		log.Warn(
			fName,
			"message", "in-memory mode: no SPIKE Keepers, not for production",
		)

		// `nil` will skip root key initialization and simply initializes an
		// in-memory backing store.
		state.Initialize(nil)
		return
	}

	// Unknown store type.
	// Better to crash, since this is likely a configuration failure.
	log.FatalLn(
		fName,
		"message",
		"invalid backend store type",
		"type", env.BackendStoreTypeVal(),
		"valid_types", "sqlite, lite, memory",
	)
}
