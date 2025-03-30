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
	http "github.com/spiffe/spike/app/keeper/internal/route/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	routing "github.com/spiffe/spike/internal/net"
)

const appName = "SPIKE Keeper"

func main() {
	log.Log().Info(appName, "msg", appName, "version", config.SpikeKeeperVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, selfSpiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	// I should be a SPIKE Keeper.
	if !auth.IsKeeper(selfSpiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", selfSpiffeid)
	}

	log.Log().Info(
		appName, "msg",
		fmt.Sprintf("Started service: %s v%s", appName, config.SpikeKeeperVersion),
	)

	if err := net.ServeWithPredicate(
		source,
		func() { routing.HandleRoute(http.Route) },
		auth.PeerCanTalkToKeeper,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
