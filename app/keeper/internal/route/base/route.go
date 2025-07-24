//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package base provides the core routing logic for the SPIKE application's
// HTTP server. It dynamically resolves incoming HTTP requests to the
// appropriate handlers based on their URL paths and methods. This package
// ensures flexibility and extensibility in supporting various API actions and
// paths within SPIKE's ecosystem.
package base

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/app/keeper/internal/route/store"
	"github.com/spiffe/spike/internal/journal"
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
//   - audit: The AuditEntry containing the client's audit information
func Route(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	return net.RouteFactory[url.ApiAction](
		url.ApiUrl(r.URL.Path),
		url.ApiAction(r.URL.Query().Get(url.KeyApiAction)),
		r.Method,
		func(a url.ApiAction, p url.ApiUrl) net.Handler {
			switch {
			// Get a contribution from SPIKE Nexus:
			case a == url.ActionDefault && p == url.SpikeKeeperUrlContribute:
				return store.RouteContribute
			// Provide your shard to SPIKE Nexus:
			case a == url.ActionDefault && p == url.SpikeKeeperUrlShard:
				return store.RouteShard
			default:
				return net.Fallback
			}
		})(w, r, audit)
}
