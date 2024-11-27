package api

import (
	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/internal/net"
	"net/url"
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

// UrlSecretMetadataGet returns the URL for getting a secret metadata.
func UrlSecretMetadataGet() string {
	u, _ := url.JoinPath(
		env.NexusApiRoot(),
		string(net.SpikeNexusUrlSecretsMetadata),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusGet))
	return u + "?" + params.Encode()
}
