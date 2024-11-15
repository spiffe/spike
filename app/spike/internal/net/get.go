//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// GetSecret retrieves a secret from SPIKE Nexus.
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
	if errors.Is(err, net.ErrNotFound) {
		return nil, nil
	}
	if errors.Is(err, net.ErrUnauthorized) {
		return nil, errors.New(
			`unauthorized. Please login first with 'spike login'`,
		)
	}

	var res reqres.SecretReadResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Join(
			errors.New("getSecret: Problem parsing response body"),
			err,
		)
	}

	return &data.Secret{Data: res.Data}, nil
}
