//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	sdkNet "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/nexus/internal/initialization"
	http "github.com/spiffe/spike/app/nexus/internal/route/base"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/out"
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
		failErr := sdkErrors.ErrStateInitializationFailed.Wrap(err)
		failErr.Msg = "failed to get SPIFFE Workload API source"
		log.FatalErr(appName, *failErr)
	}
	defer func() {
		err := spiffe.CloseSource(source)
		if err != nil {
			log.WarnErr(appName, *err)
		}
	}()

	log.Info(
		appName,
		"message", "acquired source",
		"spiffe_id", selfSPIFFEID,
	)

	// I should be SPIKE Nexus.
	if !spiffeid.IsNexus(selfSPIFFEID) {
		failErr := *sdkErrors.ErrStateInitializationFailed.Clone()
		failErr.Msg = "SPIFFE ID is not valid: " + selfSPIFFEID
		log.FatalErr(appName, failErr)
	}

	initialization.Initialize(source)

	log.Info(
		appName,
		"message", "started service",
		"version", config.NexusVersion,
	)

	// Serve the app.
	sdkNet.ServeWithRoute(
		appName,
		source,
		http.Route,
		// AllowAll, because any workload can talk to SPIKE Nexus if they
		// have a legitimate SPIFFE ID registration entry.
		// we might want to further restrict this based on environment
		// configuration maybe (for example, a predicate that checks regex
		// matching on workload SPIFFFE IDs before granting access,
		// if the matcher is not provided, AllowAll will be assumed).
		predicate.AllowAll,
		env.NexusTLSPortVal(),
	)
}
