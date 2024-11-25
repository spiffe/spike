package api

import (
	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/internal/net"
	"net/url"
)

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
