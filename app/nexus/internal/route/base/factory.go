//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/route/auth"
	"github.com/spiffe/spike/app/nexus/internal/route/store"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func factory(p net.ApiUrl, a net.SpikeNexusApiAction, m string) net.Handler {
	log.Log().Info("route.factory", "path", p, "action", a, "method", m)

	// We only accept POST requests.
	if m != http.MethodPost {
		return net.Fallback
	}

	switch {
	case a == net.ActionNexusAdminLogin && p == net.SpikeNexusUrlLogin:
		return auth.RouteAdminLogin
	case a == net.ActionNexusDefault && p == net.SpikeNexusUrlInit:
		return auth.RouteInit
	case a == net.ActionNexusCheck && p == net.SpikeNexusUrlInit:
		return auth.RouteInitCheck
	case a == net.ActionNexusDefault && p == net.SpikeNexusUrlSecrets:
		return store.RoutePutSecret
	case a == net.ActionNexusGet && p == net.SpikeNexusUrlSecrets:
		return store.RouteGetSecret
	case a == net.ActionNexusDelete && p == net.SpikeNexusUrlSecrets:
		return store.RouteDeleteSecret
	case a == net.ActionNexusUndelete && p == net.SpikeNexusUrlSecrets:
		return store.RouteUndeleteSecret
	case a == net.ActionNexusList && p == net.SpikeNexusUrlSecrets:
		return store.RouteListPaths
	// Fallback route.
	default:
		return net.Fallback
	}
}
