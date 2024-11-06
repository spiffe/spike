//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"
)

func factory(p, a, m string) handler {
	switch {
	case m == http.MethodPost && a == "" && p == urlInit:
		return routeInit
	case m == http.MethodPost && a == "" && p == urlSecrets:
		return routePostSecret
	case m == http.MethodPost && a == "get" && p == urlSecrets:
		return routeGetSecret
	case m == http.MethodPost && a == "delete" && p == urlSecrets:
		return routeDeleteSecret
	// Fallback route.
	default:
		return routeFallback
	}
}

func Route(w http.ResponseWriter, r *http.Request) {
	factory(r.URL.Path, r.URL.Query().Get("action"), r.Method)(r, w)
}
