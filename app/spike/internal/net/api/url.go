//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/url"

	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/internal/net"
)

// UrlSecretGet returns the URL for getting a secret.
func UrlSecretGet() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusGet))
	return u + "?" + params.Encode()
}

// UrlSecretPut returns the URL for putting a secret.
func UrlSecretPut() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	return u
}

// UrlSecretDelete returns the URL for deleting a secret.
func UrlSecretDelete() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusDelete))
	return u + "?" + params.Encode()
}

// UrlSecretUndelete returns the URL for undeleting a secret.
func UrlSecretUndelete() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusUndelete))
	return u + "?" + params.Encode()
}

// UrlSecretList returns the URL for listing secrets.
func UrlSecretList() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusList))
	return u + "?" + params.Encode()
}

// UrlInit returns the URL for initializing SPIKE Nexus.
func UrlInit() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlInit),
	)
	return u
}

// UrlInitState returns the URL for checking the initialization state of
// SPIKE Nexus.
func UrlInitState() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlInit),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusCheck))
	return u + "?" + params.Encode()
}

// UrlAdminLogin returns the URL for logging in as an admin.
func UrlAdminLogin() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlLogin),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusAdminLogin))
	return u + "?" + params.Encode()
}
