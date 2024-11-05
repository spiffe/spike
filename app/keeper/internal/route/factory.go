//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"
)

func factory(p, a, m string) handler {
	switch {
	case m == http.MethodPost && a == "" && p == urlKeep:
		return routeKeep
	case m == http.MethodPost && a == "read" && p == urlKeep:
		return routeShow
	// Fallback route.
	default:
		return routeFallback
	}
}
