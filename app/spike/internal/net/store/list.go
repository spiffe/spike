//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"errors"
	net2 "github.com/spiffe/spike/app/spike/internal/net"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// ListSecretKeys lists the keys of all secrets in SPIKE Nexus.
func ListSecretKeys(source *workloadapi.X509Source) ([]string, error) {
	r := reqres.SecretListRequest{}
	mr, err := json.Marshal(r)
	if err != nil {
		return []string{}, errors.Join(
			errors.New("listSecretKeys: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return []string{}, err
	}

	body, err := net.Post(client, net2.UrlSecretList(), mr)
	if errors.Is(err, net.ErrNotFound) {
		return []string{}, nil
	}
	if errors.Is(err, net.ErrUnauthorized) {
		return []string{},
			errors.New(`unauthorized. Please login first with 'spike login'`)
	}

	var res reqres.SecretListResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return []string{}, errors.Join(
			errors.New("getSecret: Problem parsing response body"),
			err,
		)
	}

	return res.Keys, nil
}
