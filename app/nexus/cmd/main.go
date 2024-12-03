//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/poll"
	"github.com/spiffe/spike/app/nexus/internal/route/handle"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
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

	err = state.Bootstrap(source)
	if err != nil {
		if errors.Is(err, state.ErrAlreadyInitialized) {
			log.Log().Info(appName,
				"msg",
				"SPIKE Nexus already initialized. Not creating a new root key.")
		} else {
			log.FatalF("Unable to initialize SPIKE Nexus state: " + err.Error())
		}
	}

	log.Log().Info(appName, "msg", "Initializing complete.",
		"has_root_key", len(state.RootKey()) > 0)

	ticker := time.NewTicker(env.PollInterval())
	defer ticker.Stop()
	go poll.Tick(ctx, source, ticker)

	log.Log().Info(appName,
		"msg", fmt.Sprintf("Started service: %s v%s",
			appName, config.NexusVersion))
	if err := net.Serve(
		source, handle.InitializeRoutes,
		auth.CanTalkToNexus,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
