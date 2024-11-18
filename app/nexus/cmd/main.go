//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"github.com/spiffe/spike/app/nexus/internal/state/base"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/handle"
	"github.com/spiffe/spike/app/nexus/internal/poll"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/internal/spiffe"
)

const appName = "SPIKE Nexus"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.AppSpiffeSource(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	trust.Authenticate(spiffeid)

	err = base.Initialize(source)
	if err != nil {
		if errors.Is(err, base.ErrAlreadyInitialized) {
			log.Log().Info(appName,
				"msg",
				"SPIKE Nexus already initialized. Not creating a new root key.")
		} else {
			log.FatalF("Unable to initialize SPIKE Nexus state: " + err.Error())
		}
	}

	log.Log().Info(appName, "msg", "Initializing complete.",
		"has_root_key", len(base.RootKey()) > 0)

	ticker := time.NewTicker(env.PollInterval())
	defer ticker.Stop()
	go poll.Tick(ctx, source, ticker)

	log.Log().Info(appName, "msg",
		"Starting service.", "version", config.NexusVersion)

	if err := net.Serve(
		source, handle.InitializeRoutes,
		auth.CanTalkToNexus,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
