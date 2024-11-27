//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package spike

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// This is a temporary POC code which will eventually evolve into an SDK.

func NexusApiRoot() string {
	p := os.Getenv("SPIKE_NEXUS_API_URL")
	if p != "" {
		return p
	}
	return "https://localhost:8553"
}

func UrlSecretGet() string {
	u, _ := url.JoinPath(
		NexusApiRoot(),
		string(net.SpikeNexusUrlSecrets),
	)
	params := url.Values{}
	params.Add(net.KeyApiAction, string(net.ActionNexusGet))
	return u + "?" + params.Encode()
}

func GetSecret(source *workloadapi.X509Source,
	path string, version int) (*data.Secret, error) {
	r := reqres.SecretReadRequest{
		Path:    path,
		Version: version,
	}

	mr, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Join(
			errors.New("getSecret: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return nil, err
	}

	body, err := net.Post(client, UrlSecretGet(), mr)
	if err != nil {
		if errors.Is(err, net.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var res reqres.SecretReadResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Join(
			errors.New("getSecret: Problem parsing response body"),
			err,
		)
	}
	if res.Err != "" {
		return nil, errors.New(string(res.Err))
	}

	return &data.Secret{Data: res.Data}, nil
}
