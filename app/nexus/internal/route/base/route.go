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

	"github.com/spiffe/spike/app/nexus/internal/env"
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
			emptyRootKey := state.RootKeyZero()
			inMemoryMode := env.BackendStoreType() == env.Memory
			fullMode := env.BackendStoreType() != env.Lite
			emergencyAction := p == url.SpikeNexusUrlOperatorRecover ||
				p == url.SpikeNexusUrlOperatorRestore
			canBypassRootKeyValidation := inMemoryMode || emergencyAction
			rootKeyValidationRequired := !canBypassRootKeyValidation

			if rootKeyValidationRequired && emptyRootKey {
				return net.NotReady
			}

			switch {
			case a == url.ActionDefault && p == url.SpikeNexusUrlSecrets && fullMode:
				return secret.RoutePutSecret
			case a == url.ActionGet && p == url.SpikeNexusUrlSecrets && fullMode:
				return secret.RouteGetSecret
			case a == url.ActionDelete && p == url.SpikeNexusUrlSecrets && fullMode:
				return secret.RouteDeleteSecret
			case a == url.ActionUndelete && p == url.SpikeNexusUrlSecrets && fullMode:
				return secret.RouteUndeleteSecret
			case a == url.ActionList && p == url.SpikeNexusUrlSecrets && fullMode:
				return secret.RouteListPaths
			case a == url.ActionDefault && p == url.SpikeNexusUrlPolicy && fullMode:
				return policy.RoutePutPolicy
			case a == url.ActionGet && p == url.SpikeNexusUrlPolicy && fullMode:
				return policy.RouteGetPolicy
			case a == url.ActionDelete && p == url.SpikeNexusUrlPolicy && fullMode:
				return policy.RouteDeletePolicy
			case a == url.ActionList && p == url.SpikeNexusUrlPolicy && fullMode:
				return policy.RouteListPolicies
			case a == url.ActionGet && p == url.SpikeNexusUrlSecretsMetadata && fullMode:
				return secret.RouteGetSecretMetadata
			case a == url.ActionDefault && p == url.SpikeNexusUrlOperatorRestore:
				return operator.RouteRestore
			case a == url.ActionDefault && p == url.SpikeNexusUrlOperatorRecover:
				return operator.RouteRecover
			// FIXME move constant
			case a == url.ActionDefault && p == "/v1/encrypt" && !fullMode:
				return secret.RouteEncrypt
			case a == url.ActionDefault && p == "/v1/decrypt" && !fullMode:
				return secret.RouteDecrypt
			default:
				return net.Fallback
			}
		})(w, r, a)
}
