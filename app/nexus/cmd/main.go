//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/out"

	"github.com/spiffe/spike/app/nexus/internal/initialization"
	"github.com/spiffe/spike/app/nexus/internal/net"
	"github.com/spiffe/spike/internal/config"
)

const appName = "SPIKE Nexus"

func main() {
	out.Preamble(appName, config.NexusVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Log().Info(
		appName,
		"message", "SPIFFE trust root: "+env.TrustRootVal(),
	)

	source, selfSPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(appName, "message", "failed to get source", "err", err.Error())
	}
	defer spiffe.CloseSource(source)

	log.Log().Info(appName, "message", "self.spiffeid: "+selfSPIFFEID)

	// I should be SPIKE Nexus.
	if !spiffeid.IsNexus(selfSPIFFEID) {
		log.FatalLn(appName,
			"message",
			"Authenticate: SPIFFE ID is not valid",
			"spiffeid", selfSPIFFEID)
	}

	initialization.Initialize(source)

	log.Log().Info(appName, "message", fmt.Sprintf(
		"Started service: %s v%s",
		appName, config.NexusVersion),
	)

	// Start the server:
	net.Serve(appName, source)
}
