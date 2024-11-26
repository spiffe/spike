package api

import (
	"net/url"

	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/internal/net"
)

// UrlPolicyCreate returns the URL for creating a policy.
func UrlPolicyCreate() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	return u
}

// UrlPolicyList returns the URL for listing policies.
func UrlPolicyList() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusList))
	return u + "?" + params.Encode()
}

// UrlPolicyDelete returns the URL for deleting a policy.
func UrlPolicyDelete() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusDelete))
	return u + "?" + params.Encode()
}

// UrlPolicyGet returns the URL for getting a policy.
func UrlPolicyGet() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusGet))
	return u + "?" + params.Encode()
}
