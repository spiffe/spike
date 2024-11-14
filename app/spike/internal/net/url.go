//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/url"

	"github.com/spiffe/spike/app/spike/internal/env"
)

func UrlSecretGet() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/secrets?action=get")
	return u
}

func UrlSecretPut() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/secrets")
	return u
}

func UrlSecretDelete() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/secrets?action=delete")
	return u
}

func UrlSecretUndelete() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/secrets?action=undelete")
	return u
}

func UrlSecretList() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/secrets?action=list")
	return u
}

func UrlInit() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/init")
	return u
}

func UrlInitState() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/init?action=check")
	return u
}

func UrlAdminLogin() string {
	u, _ := url.JoinPath(env.NexusApiRoot(), "/v1/login?action=admin")
	return u
}
