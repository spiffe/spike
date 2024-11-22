//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/url"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/net"
)

// UrlKeeperRead returns the full URL for the SPIKE Keeper API read endpoint.
// The URL is constructed by joining the base Keep API root path with
// "/v1/keep?action=read". Any errors from url joining are ignored.
//
// Returns:
//   - string: The complete URL for the read endpoint
func UrlKeeperRead() string {
	u, _ := url.JoinPath(env.KeepApiRoot(), string(net.SpikeKeeperUrlKeep))
	params := url.Values{}
	params.Add("action", "read")
	return u + "?" + params.Encode()
}

// UrlKeeperWrite returns the hardcoded URL for the SPIKE Keeper API write
// endpoint.
//
// Returns:
//   - string: The complete URL for the write endpoint
func UrlKeeperWrite() string {
	u, _ := url.JoinPath(env.KeepApiRoot(), string(net.SpikeKeeperUrlKeep))
	return u
}
