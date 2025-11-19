//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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

	log.Info(
		appName,
		"message", "starting",
		"spiffe_trust_root", env.TrustRootVal(),
	)

	source, selfSPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		failErr := sdkErrors.ErrInitializationFailed.Wrap(err)
		failErr.Msg = "failed to get SPIFFE Workload API source"
		log.FatalErr(appName, *failErr)
	}
	defer spiffe.CloseSource(source)

	log.Info(
		appName,
		"message", "acquired source",
		"spiffe_id", selfSPIFFEID,
	)

	// I should be SPIKE Nexus.
	if !spiffeid.IsNexus(selfSPIFFEID) {
		failErr := sdkErrors.ErrInitializationFailed
		failErr.Msg = "SPIFFE ID is not valid: " + selfSPIFFEID
		log.FatalErr(appName, *failErr)
	}

	initialization.Initialize(source)

	log.Info(
		appName,
		"message", "started service",
		"version", config.NexusVersion,
	)
	// Start the server:
	net.Serve(appName, source)
}
