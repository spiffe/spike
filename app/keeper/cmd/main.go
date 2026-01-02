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
	http "github.com/spiffe/spike/app/keeper/internal/route/base"
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
		log.FatalErr(appName, *sdkErrors.ErrStateInitializationFailed.Wrap(err))
	}
	defer func() {
		closeErr := spiffe.CloseSource(source)
		if closeErr != nil {
			log.WarnErr(
				appName, *sdkErrors.ErrSPIFFEFailedToCloseX509Source.Wrap(closeErr),
			)
		}
	}()

	// I should be a SPIKE Keeper.
	if !spiffeid.IsKeeper(selfSPIFFEID) {
		failErr := *sdkErrors.ErrStateInitializationFailed.Clone()
		failErr.Msg = "SPIFFE ID is not valid: " + selfSPIFFEID
		log.FatalErr(appName, failErr)
	}

	log.Info(
		appName,
		"message", "started service",
		"version", config.KeeperVersion,
	)

	// Serve the app.
	sdkNet.ServeWithRoute(
		appName,
		source,
		http.Route,
		predicate.AllowKeeperPeer,
		env.KeeperTLSPortVal(),
	)
}
