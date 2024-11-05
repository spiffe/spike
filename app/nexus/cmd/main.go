//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/handle"
	"github.com/spiffe/spike/app/nexus/internal/poll"
	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/config"
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

	if !config.IsNexus(spiffeid) {
		log.Fatalf("SPIFFE ID %s is not valid.\n", spiffeid)
	}

	err = state.Initialize(source)
	if err != nil {
		if errors.Is(err, state.ErrAlreadyInitialized) {
			log.Println("SPIKE Nexus already initialized. Not creating a new root key.")
		} else {
			log.Fatalf("Unable to initialize SPIKE Nexus state: " + err.Error())
		}
	}

	log.Println("SPIKE Nexus: Initializing complete. Has root key?", len(state.RootKey()) > 0)

	ticker := time.NewTicker(10 * time.Second) // for now; will be configurable.
	defer ticker.Stop()
	go poll.Tick(ctx, source, ticker)

	log.Printf("Started service: %s v%s\n", appName, config.NexusVersion)
	if err := net.Serve(
		source, handle.InitializeRoutes,
		config.CanTalkToNexus,
		env.TlsPort(),
	); err != nil {
		log.Fatalf("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
