//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"github.com/spiffe/spike/app/nexus/internal/env"
	"net/url"
)

// UrlKeepRead returns the full URL for the SPIKE Keeper API read endpoint.
// The URL is constructed by joining the base Keep API root path with
// "/v1/keep?action=read". Any errors from url joining are ignored.
//
// Returns:
//   - string: The complete URL for the read endpoint
func UrlKeepRead() string {
	u, _ := url.JoinPath(env.KeepApiRoot(), "/v1/keep?action=read")
	return u
}

// UrlKeepWrite returns the hardcoded URL for the SPIKE Keeper API write endpoint.
//
// Returns:
//   - string: The complete URL for the write endpoint
func UrlKeepWrite() string {
	u, _ := url.JoinPath(env.KeepApiRoot(), "/v1/keep")
	return u
}
