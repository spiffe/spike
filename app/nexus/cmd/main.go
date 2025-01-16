//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/poll"
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

	ticker := time.NewTicker(env.PollInterval())
	defer ticker.Stop()

	// Waits until SPIKE Nexus fetches adequate trust material from
	// SPIKE Keepers to compute a root key for the backing store.

	// TODO: since this is not polling anymore, `Tick` is not the right name
	// for it.
	poll.Tick(source)
	// TODO: Later: for in-memory backing store, we can bypass this poll and
	// initialize the backing store with some dummy 32byte random data.

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
