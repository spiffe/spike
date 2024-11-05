//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"log"

	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/app/keeper/internal/handle"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/internal/spiffe"
)

const appName = "SPIKE Keeper"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.AppSpiffeSource(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer spiffe.CloseSource(source)

	if !config.IsKeeper(spiffeid) {
		log.Fatalf("SPIFFE ID %s is not valid.\n", spiffeid)
	}

	log.Printf("Started service: %s v%s\n", appName, config.KeeperVersion)
	if err := net.Serve(
		source, handle.InitializeRoutes,
		config.CanTalkToKeeper,
		env.TlsPort(),
	); err != nil {
		log.Fatalf("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
