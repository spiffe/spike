//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/keeper/internal/env"
	http "github.com/spiffe/spike/app/keeper/internal/route/base"
	"github.com/spiffe/spike/internal/log"
	routing "github.com/spiffe/spike/internal/net"
)

func serve(source *workloadapi.X509Source) {
	if err := net.ServeWithPredicate(
		source,
		func() { routing.HandleRoute(http.Route) },
		func(peerSpiffeId string) bool {
			// Only SPIKE Nexus can talk to SPIKE Keeper:
			return spiffeid.PeerCanTalkToKeeper(env.TrustRootForNexus(), peerSpiffeId)
		},
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
