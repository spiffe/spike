package api

import (
	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/internal/net"
	"net/url"
)

func UrlPolicyCreate() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	return u
}

func UrlPolicyList() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusList))
	return u + "?" + params.Encode()
}

func UrlPolicyDelete() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusDelete))
	return u + "?" + params.Encode()
}

func UrlPolicyGet() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlPolicy),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusGet))
	return u + "?" + params.Encode()
}
