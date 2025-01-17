//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/crypto"
)

func Initialize(source *workloadapi.X509Source) {
	requireBootstrapping := env.BackendStoreType() == env.Sqlite
	if requireBootstrapping {
		bootstrap(source)

		// If bootstrapping is successful, start a background process to
		// periodically sync shards.
		go sendShardsPeriodically(source)

		return
	}

	// For "in-memory" backing stores, we don't need bootstrapping.
	// Initialize the store with a random seed instead.
	seed, err := crypto.Aes256Seed()
	if err != nil {
		log.Fatal(err.Error())
	}
	state.Initialize(seed)
}
