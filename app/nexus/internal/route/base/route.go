//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/journal"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// Route handles all incoming HTTP requests by dynamically selecting and
// executing the appropriate handler based on the request path and HTTP method.
// It uses a factory function to create the specific handler for the given URL
// path and HTTP method combination.
//
// Parameters:
//   - w: The HTTP ResponseWriter to write the response to
//   - r: The HTTP Request containing the client's request details
func Route(
	w http.ResponseWriter, r *http.Request, a *journal.AuditEntry,
) *sdkErrors.SDKError {
	return net.RouteFactory[url.APIAction](
		url.APIURL(r.URL.Path),
		url.APIAction(r.URL.Query().Get(url.KeyAPIAction)),
		r.Method,
		func(a url.APIAction, p url.APIURL) net.Handler {
			// Lite: requires the root key.
			// SQLite: requires the root key.
			// Memory: does not require the root key.

			emptyRootKey := state.RootKeyZero()
			inMemoryMode := env.BackendStoreTypeVal() == env.Memory
			hasBackingStore := env.BackendStoreTypeVal() != env.Lite
			emergencyAction := p == url.NexusOperatorRecover ||
				p == url.NexusOperatorRestore
			rootKeyValidationRequired := !inMemoryMode && !emergencyAction

			if rootKeyValidationRequired && emptyRootKey {
				return net.NotReady
			}

			if hasBackingStore {
				return routeWithBackingStore(a, p)
			}

			// No backing store: We cannot store or retrieve secrets
			// or policies directly.
			// SPIKE is effectively a "crypto as a service" now.
			return routeWithNoBackingStore(a, p)
		})(w, r, a)
}
