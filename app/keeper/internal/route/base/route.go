//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/route/store"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// Route handles all incoming HTTP requests by dynamically selecting and
// executing the appropriate handler based on the request path and HTTP method.
// It uses a factory function to create the specific handler for the given URL
// path and HTTP method combination.
//
// Parameters:
//   - w: The HTTP ResponseWriter to write the response to
//   - r: The HTTP Request containing the client's request details
func Route(w http.ResponseWriter, r *http.Request, a *log.AuditEntry) error {
	return net.RouteFactory[net.SpikeKeeperApiAction](
		net.ApiUrl(r.URL.Path),
		net.SpikeKeeperApiAction(r.URL.Query().Get(net.KeyApiAction)),
		r.Method,
		func(a net.SpikeKeeperApiAction, p net.ApiUrl) net.Handler {
			switch {
			case a == net.ActionKeeperDefault && p == net.SpikeKeeperUrlKeep:
				return store.RouteKeep
			case a == net.ActionKeeperRead && p == net.SpikeKeeperUrlKeep:
				return store.RouteShow
			case a == net.ActionKeeperDefault && p == net.SpikeKeeperUrlContribute:
				return store.RouteContribute
			case a == net.ActionKeeperDefault && p == net.SpikeKeeperUrlShard:
				return store.RouteShard
			case a == net.ActionKeeperDefault && p == net.SpikeKeeperUrlStatus:
				return store.RouteStatus
			default:
				return net.Fallback
			}
		})(w, r, a)
}
