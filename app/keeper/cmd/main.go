//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/keeper/internal/net"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/out"
)

const appName = "SPIKE Keeper"

func main() {
	out.Preamble(appName, config.KeeperVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, selfSPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	// I should be a SPIKE Keeper.
	if !spiffeid.IsKeeper(selfSPIFFEID) {
		log.FatalLn(
			appName,
			"message", "SPIFFE ID is not valid",
			"spiffeid", selfSPIFFEID,
		)
	}

	log.Log().Info(
		appName,
		"message", "started service",
		"app", appName,
		"version", config.KeeperVersion,
	)

	// Serve the app.
	net.Serve(appName, source)
}
