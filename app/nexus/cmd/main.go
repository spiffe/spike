//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization"
	http "github.com/spiffe/spike/app/nexus/internal/route/base"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	routing "github.com/spiffe/spike/internal/net"
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

	initialization.Initialize(source)

	log.Log().Info(appName, "msg", fmt.Sprintf(
		"Started service: %s v%s",
		appName, config.NexusVersion),
	)

	if err := net.Serve(
		source,
		func() { routing.HandleRoute(http.Route) },
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
