//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package base contains the fundamental building blocks and core functions
// for handling HTTP requests in the SPIKE Nexus application. It provides
// the routing logic to map API actions and URL paths to their respective
// handlers while ensuring seamless request processing and response generation.
// This package serves as a central point for managing incoming API calls
// and delegating them to the correct functional units based on specified rules.
package base

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/app/nexus/internal/route/acl/policy"
	"github.com/spiffe/spike/app/nexus/internal/route/operator"
	"github.com/spiffe/spike/app/nexus/internal/route/secret"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
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
func Route(
	w http.ResponseWriter, r *http.Request, a *log.AuditEntry,
) error {
	return net.RouteFactory[url.ApiAction](
		url.ApiUrl(r.URL.Path),
		url.ApiAction(r.URL.Query().Get(url.KeyApiAction)),
		r.Method,
		func(a url.ApiAction, p url.ApiUrl) net.Handler {
			rk := state.RootKey()

			emptyRootKey := len(rk) == 0
			emergencyAction := p == url.SpikeNexusUrlOperatorRecover ||
				p == url.SpikeNexusUrlOperatorRestore
			if emptyRootKey && !emergencyAction {
				return net.NotReady
			}

			switch {
			case a == url.ActionDefault && p == url.SpikeNexusUrlSecrets:
				return secret.RoutePutSecret
			case a == url.ActionGet && p == url.SpikeNexusUrlSecrets:
				return secret.RouteGetSecret
			case a == url.ActionDelete && p == url.SpikeNexusUrlSecrets:
				return secret.RouteDeleteSecret
			case a == url.ActionUndelete && p == url.SpikeNexusUrlSecrets:
				return secret.RouteUndeleteSecret
			case a == url.ActionList && p == url.SpikeNexusUrlSecrets:
				return secret.RouteListPaths
			case a == url.ActionDefault && p == url.SpikeNexusUrlPolicy:
				return policy.RoutePutPolicy
			case a == url.ActionGet && p == url.SpikeNexusUrlPolicy:
				return policy.RouteGetPolicy
			case a == url.ActionDelete && p == url.SpikeNexusUrlPolicy:
				return policy.RouteDeletePolicy
			case a == url.ActionList && p == url.SpikeNexusUrlPolicy:
				return policy.RouteListPolicies
			case a == url.ActionGet && p == url.SpikeNexusUrlSecretsMetadata:
				return secret.RouteGetSecretMetadata
			case a == url.ActionDefault && p == url.SpikeNexusUrlOperatorRestore:
				return operator.RouteRestore
			case a == url.ActionDefault && p == url.SpikeNexusUrlOperatorRecover:
				return operator.RouteRecover
			default:
				return net.Fallback
			}
		})(w, r, a)
}
