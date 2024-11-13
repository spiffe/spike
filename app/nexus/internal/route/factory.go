//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/internal/net"
)

func factory(p, a, m string) net.Handler {
	switch {
	case m == http.MethodPost && a == "admin" && p == urlLogin:
		return routeAdminLogin
	case m == http.MethodPost && a == "" && p == urlInit:
		return routeInit
	case m == http.MethodPost && a == "check" && p == urlInit:
		return routeInitCheck
	case m == http.MethodPost && a == "" && p == urlSecrets:
		return routePutSecret
	case m == http.MethodPost && a == "get" && p == urlSecrets:
		return routeGetSecret
	case m == http.MethodPost && a == "delete" && p == urlSecrets:
		return routeDeleteSecret
	case m == http.MethodPost && a == "undelete" && p == urlSecrets:
		return routeUndeleteSecret
	case m == http.MethodPost && a == "list" && p == urlSecrets:
		return routeListPaths
	// Fallback route.
	default:
		return net.Fallback
	}
}
