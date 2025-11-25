//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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
		log.FatalErr(appName, *sdkErrors.ErrStateInitializationFailed.Wrap(err))
	}
	defer spiffe.CloseSource(source)

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
	net.Serve(appName, source)
}
