//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/keeper/internal/route/store"
	"net/http"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func factory(p net.ApiUrl, a net.SpikeKeeperApiAction, m string) net.Handler {
	log.Log().Info("route.factory", "path", p, "action", a, "method", m)

	// We only accept POST requests -- See ADR-0012.
	if m != http.MethodPost {
		return net.Fallback
	}

	switch {
	case a == net.ActionKeeperDefault && p == net.SpikeKeeperUrlKeep:
		return store.RouteKeep
	case a == net.ActionKeeperRead && p == net.SpikeKeeperUrlKeep:
		return store.RouteShow
	// Fallback route.
	default:
		return net.Fallback
	}
}
