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

// Initialize initializes the SPIKE Nexus backing store based on the configured
// backend store type. The function handles two initialization modes:
//
// 1. Keeper-based initialization (for SQLite and Lite backend types):
//   - Initializes the backing store from SPIKE Keeper instances
//   - Starts a background process for periodic shard synchronization
//
// 2. In-memory initialization (for other backend types):
//   - Generates a cryptographically secure 32-byte root key
//   - Initializes the internal state with the generated root key
//   - Securely zeros out the root key from memory after use
//
// The source parameter provides the X.509 certificates and private keys
// needed for SPIFFE-based authentication when communicating with SPIKE Keepers.
//
// Security considerations:
//   - Root keys are generated using crypto/rand for cryptographic security
//   - Memory is explicitly zeroed after use to prevent key material leakage
//   - The root key is passed by reference to avoid inadvertent copying
//
// Note: This function will call log.Fatal and terminate the program if
// cryptographically secure random number generation fails.
func Initialize(source *workloadapi.X509Source) {
	const fName = "Initialize"

	requireKeepersToBootstrap := env.BackendStoreType() == env.Sqlite ||
		env.BackendStoreType() == env.Lite

	if requireKeepersToBootstrap {
		// Initialize the backing store from SPIKE Keeper instances.
		// This is only required when the SPIKE Nexus needs bootstrapping.
		// For modes where bootstrapping is not required (such as in-memory mode),
		//SPIKE Nexus should be initialized internally.
		recovery.InitializeBackingStoreFromKeepers(source)

		// Lazy evaluation in a loop:
		// If bootstrapping is successful, start a background process to
		// periodically sync shards.
		go recovery.SendShardsPeriodically(source)

		return
	}

	log.Log().Info(fName, "message", "Will not use SPIKE Keepers.")

	// Security: Use a static byte array and pass it as a pointer to avoid
	// inadvertent pass-by-value copying / memory allocation.
	var rootKey [32]byte
	// Security: Zero-out rootKey after persisted internally.
	defer func() {
		// Note: Each function must zero-out ONLY the items it has created.
		// If it is borrowing an item by reference, it must not zero-out the item
		// and let the owner zero-out the item.
		mem.ClearRawBytes(&rootKey)
	}()

	if _, err := rand.Read(rootKey[:]); err != nil {
		log.Fatal(err.Error())
	}

	state.Initialize(&rootKey)
}
