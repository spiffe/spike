//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"crypto/rand"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// Initialize sets up the system state based on the backend store type.
// For SQLite backends, it performs bootstrapping and starts periodic shard
// synchronization.  For in-memory stores, it initializes with a random seed.
//
// Parameters:
//   - source: *workloadapi.X509Source for SPIFFE workload API authentication
//
// The function will:
//   - Bootstrap from SQLite if using SQLite backend
//   - Start periodic shard syncing for SQLite backend
//   - Initialize with random seed for in-memory backend
//
// Panics if random seed generation fails for in-memory stores.
func Initialize(source *workloadapi.X509Source) {
	const fName = "Initialize"
	requireBootstrapping := env.BackendStoreType() == env.Sqlite || env.BackendStoreType() == env.Lite
	if requireBootstrapping {
		log.Log().Info(fName, "message", "Backend store requires bootstrapping")

		// Try bootstrapping in a loop.
		go bootstrap(source)

		// Lazy evaluation in a loop:
		// If bootstrapping is successful, start a background process to
		// periodically sync shards.
		go recovery.SendShardsPeriodically(source)

		return
	}

	log.Log().Info(fName, "message", "Backend store does not require bootstrapping")

	// Security: Use a static byte array and pass it as a pointer to avoid
	// inadvertent pass-by-value copying / memory allocation.
	var seed [32]byte

	// Security: Zero-out seed after use.
	defer func() {
		// Note: Each function must zero-out ONLY the items it has created.
		// If it is borrowing an item by reference, it must not zero-out the item
		// and let the owner zero-out the item.
		//
		// For example, `seed` should be reset here,
		// but not in `state.Initialize()`.
		mem.ClearRawBytes(&seed)
	}()

	if _, err := rand.Read(seed[:]); err != nil {
		log.Fatal(err.Error())
	}

	state.Initialize(&seed)
}
