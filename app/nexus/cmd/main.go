//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization"
	"github.com/spiffe/spike/app/nexus/internal/route/handle"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Nexus"

// TODO:
// 1. Ensure that SPIKE Nexus can successfully recover.
// 2. Have a /health API for keepers. If SPIKE Nexus cannot get a healthy
//      response from a keeper, it should run a background process to re-send
//      shards. The /health API should be more frequent than the periodic shard
//      syncing. The health API will report whether the keeper has a shard too
//      so that nexus can decide to re-seed it or not.
//
// 3. Doomsday recovery:
//    3.1. Have a /recover endpoint that dumps two shards as a response.
//    3.2. Have SPIKE Pilot save those shards in two separate files under ~/.spike
//    3.3. Have a /restore endpoint on SPIKE Nexus that accepts one shard at
//     a time.
//    3.4. When restore has enough shards, it recomputes the root key, updates
//     internal state, starts a new bootstrapping flow.
//    other notes:
//      - Keepers are designed to not know anything about Nexus or themselves
//        this means, we can send them data, but they wouldn't know about their
//        actual state (i.e. whether they are at day zero and that's why they
//        don't have a shard; or they crashed and they don't have a shard.
//        only thing they have is a health endpoint. and nexus is responsible
//        to reseed them, if they report unhealthy.

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

	initialization.Initialize(source)

	log.Log().Info(appName, "msg", fmt.Sprintf(
		"Started service: %s v%s",
		appName, config.NexusVersion),
	)

	if err := net.Serve(
		source, handle.InitializeRoutes,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
