//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/keeper/internal/env"
	api "github.com/spiffe/spike/app/keeper/internal/net"
	"github.com/spiffe/spike/app/keeper/internal/route/handle"
	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/app/keeper/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Keeper"

func main() {
	log.Log().Info(appName, "msg", appName, "version", config.KeeperVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	trust.Authenticate(spiffeid)

	// TODO: log.Fatalf instead of panicking.

	// 1. Load State
	keeperState := state.ReadAppState()
	if keeperState == state.AppStateError {
		panic("Error reading state file")
	}
	if keeperState == state.AppStateNotReady {
		log.Log().Info(appName, "msg", "Not ready. Will send shards")

		go api.Contribute(source)
		go state.WaitForShards()
	}
	if keeperState == state.AppStateReady {
		// TODO: implement this case
		// Keeper should query its peers to recompute its shard.
		panic("I started, but I don't know what to do.")
	}

	log.Log().Info(appName, "msg", fmt.Sprintf("Started service: %s v%s",
		appName, config.KeeperVersion))
	if err := net.ServeWithPredicate(
		source, handle.InitializeRoutes,
		auth.CanTalkToKeeper,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
