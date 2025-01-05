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

	keeperState := state.ReadAppState()

	if keeperState == state.AppStateError {
		log.FatalLn(
			"SPIKE Keeper is in ERROR state. Manual intervention required.",
		)
	}

	if keeperState == state.AppStateNotReady {
		log.Log().Info(appName,
			"msg", "SPIKE Keeper is not ready. Will send shards")

		// go api.Contribute(source)
		// go state.WaitForShards()
	}

	if keeperState == state.AppStateReady ||
		keeperState == state.AppStateRecovering {
		// TODO: implement this case
		// 1. Transition to a RECOVERING state, if not done already
		// 2. Contact peers to recompute shard.
		// 3. Try forever.
		// 4. If something is irrevocably irrecoverable transition to ERROR state.
		// 5. When everything is back to normal, transition to READY state.
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
