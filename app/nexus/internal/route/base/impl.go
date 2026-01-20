//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/nexus/internal/route/acl/policy"
	"github.com/spiffe/spike/app/nexus/internal/route/bootstrap"
	"github.com/spiffe/spike/app/nexus/internal/route/cipher"
	"github.com/spiffe/spike/app/nexus/internal/route/operator"
	"github.com/spiffe/spike/app/nexus/internal/route/secret"
)

// routeWithBackingStore maps API actions and URLs to their corresponding
// handlers when the backing store is initialized and available.
//
// This function routes requests to handlers that require an initialized
// backing store, including secret operations, policy management, metadata
// queries, operator functions, cipher operations, and bootstrap verification.
//
// Parameters:
//   - a: The API action to perform (e.g., Get, Delete, List)
//   - p: The API URL path identifier
//
// Returns:
//   - net.Handler: The appropriate handler for the given action and URL,
//     or net.Fallback if no matching route is found
func routeWithBackingStore(a url.APIAction, p url.APIURL) net.Handler {
	switch {
	case a == url.ActionDefault && p == url.NexusSecrets:
		return secret.RoutePutSecret
	case a == url.ActionGet && p == url.NexusSecrets:
		return secret.RouteGetSecret
	case a == url.ActionDelete && p == url.NexusSecrets:
		return secret.RouteDeleteSecret
	case a == url.ActionUndelete && p == url.NexusSecrets:
		return secret.RouteUndeleteSecret
	case a == url.ActionList && p == url.NexusSecrets:
		return secret.RouteListPaths
	case a == url.ActionDefault && p == url.NexusPolicy:
		return policy.RoutePutPolicy
	case a == url.ActionGet && p == url.NexusPolicy:
		return policy.RouteGetPolicy
	case a == url.ActionDelete && p == url.NexusPolicy:
		return policy.RouteDeletePolicy
	case a == url.ActionList && p == url.NexusPolicy:
		return policy.RouteListPolicies
	case a == url.ActionGet && p == url.NexusSecretsMetadata:
		return secret.RouteGetSecretMetadata
	case a == url.ActionDefault && p == url.NexusOperatorRestore:
		return operator.RouteRestore
	case a == url.ActionDefault && p == url.NexusOperatorRecover:
		return operator.RouteRecover
	case a == url.ActionDefault && p == url.NexusCipherEncrypt:
		return cipher.RouteEncrypt
	case a == url.ActionDefault && p == url.NexusCipherDecrypt:
		return cipher.RouteDecrypt
	case a == url.ActionDefault && p == url.NexusBootstrapVerify:
		return bootstrap.RouteVerify
	default:
		return net.Fallback
	}
}

// routeWithNoBackingStore maps API actions and URLs to their corresponding
// handlers when the backing store is not yet initialized.
//
// This function provides limited routing for operations that can function
// without an initialized backing store. Only operator recovery/restore and
// cipher operations are available in this mode. All other requests are
// routed to the fallback handler.
//
// Parameters:
//   - a: The API action to perform
//   - p: The API URL path identifier
//
// Returns:
//   - net.Handler: The appropriate handler for the given action and URL,
//     or net.Fallback if the operation requires a backing store
func routeWithNoBackingStore(a url.APIAction, p url.APIURL) net.Handler {
	switch {
	case a == url.ActionDefault && p == url.NexusOperatorRecover:
		return operator.RouteRecover
	case a == url.ActionDefault && p == url.NexusOperatorRestore:
		return operator.RouteRestore
	case a == url.ActionDefault && p == url.NexusCipherEncrypt:
		return cipher.RouteEncrypt
	case a == url.ActionDefault && p == url.NexusCipherDecrypt:
		return cipher.RouteDecrypt
	default:
		return net.Fallback
	}
}
