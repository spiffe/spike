//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func factory(p string, a net.SpikeKeeperApiAction, m string) net.Handler {
	log.Log().Info("route.factory", "path", p, "action", a, "method", m)

	switch {
	case m == http.MethodPost && a == net.ActionKeeperDefault && p == urlKeep:
		return routeKeep
	case m == http.MethodPost && a == net.ActionKeeperRead && p == urlKeep:
		return routeShow
	// Fallback route.
	default:
		return net.Fallback
	}
}
