//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/pkg/crypto"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization"
	"github.com/spiffe/spike/app/nexus/internal/route/handle"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Nexus"

func main() {
	log.Log().Info(appName, "msg", appName, "version", config.NexusVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	trust.Authenticate(spiffeid)

	requireBootstrapping := env.BackendStoreType() == env.Sqlite
	if requireBootstrapping {
		initialization.Bootstrap(source)

		// If bootstrapping is successful, start a background process to
		// periodically sync shards.
		go initialization.SendShardsPeriodically(source)
	} else {
		// For "in-memory" backing stores, we don't need bootstrapping.
		// Initialize the store with a random seed instead.
		seed, err := crypto.Aes256Seed()
		if err != nil {
			log.Fatal(err.Error())
		}
		state.Initialize(seed)
	}

	log.Log().Info(appName,
		"msg", fmt.Sprintf("Started service: %s v%s",
			appName, config.NexusVersion))

	if err := net.Serve(
		source, handle.InitializeRoutes,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
