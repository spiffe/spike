//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package acl

import (
	"encoding/json"
	"errors"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/api"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

func ListPolicies(source *workloadapi.X509Source) (*[]data.Policy, error) {
	r := reqres.PolicyListRequest{}
	mr, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Join(
			errors.New("listPolicies: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return nil, err
	}

	body, err := net.Post(client, api.UrlPolicyList(), mr)
	if err != nil {
		if errors.Is(err, net.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var res reqres.PolicyListResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Join(
			errors.New("getPolicy: Problem parsing response body"),
			err,
		)
	}
	if res.Err != "" {
		return nil, errors.New(string(res.Err))
	}

	return &res.Policies, nil
}
