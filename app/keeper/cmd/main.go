//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/keeper/internal/env"
	http "github.com/spiffe/spike/app/keeper/internal/route/base"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	routing "github.com/spiffe/spike/internal/net"
)

const appName = "SPIKE Keeper"

func main() {
	fmt.Printf(
		"%s v%s. | SPIKE secures your secrets with SPIFFE. | LOG LEVEL: %s\n",
		appName, config.SpikeKeeperVersion, log.Level(),
	)

	if mem.Lock() {
		log.Log().Info(appName, "msg", "Successfully locked memory.")
	} else {
		log.Log().Info(appName, "msg", "Memory is not locked. Please disable swap.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, selfSpiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	// I should be a SPIKE Keeper.
	if !spiffeid.IsKeeper(selfSpiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", selfSpiffeid)
	}

	log.Log().Info(
		appName, "msg",
		fmt.Sprintf("Started service: %s v%s", appName, config.SpikeKeeperVersion),
	)

	if err := net.ServeWithPredicate(
		source,
		func() { routing.HandleRoute(http.Route) },
		spiffeid.PeerCanTalkToKeeper,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
