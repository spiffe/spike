//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func factory(p string, a net.SpikeNexusApiAction, m string) net.Handler {
	log.Log().Info("route.factory", "path", p, "action", a, "method", m)

	switch {
	case m == http.MethodPost && a == net.ActionNexusAdminLogin && p == urlLogin:
		return routeAdminLogin
	case m == http.MethodPost && a == net.ActionNexusDefault && p == urlInit:
		return routeInit
	case m == http.MethodPost && a == net.ActionNexusCheck && p == urlInit:
		return routeInitCheck
	case m == http.MethodPost && a == net.ActionNexusDefault && p == urlSecrets:
		return routePutSecret
	case m == http.MethodPost && a == net.ActionNexusGet && p == urlSecrets:
		return routeGetSecret
	case m == http.MethodPost && a == net.ActionNexusDelete && p == urlSecrets:
		return routeDeleteSecret
	case m == http.MethodPost && a == net.ActionNexusUndelete && p == urlSecrets:
		return routeUndeleteSecret
	case m == http.MethodPost && a == net.ActionNexusList && p == urlSecrets:
		return routeListPaths
	// Fallback route.
	default:
		return net.Fallback
	}
}
