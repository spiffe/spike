//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/app/nexus/internal/route/acl/policy"
	"github.com/spiffe/spike/app/nexus/internal/route/cipher"
	"github.com/spiffe/spike/app/nexus/internal/route/operator"
	"github.com/spiffe/spike/app/nexus/internal/route/secret"
	"github.com/spiffe/spike/internal/net"
)

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
	default:
		return net.Fallback
	}
}

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
