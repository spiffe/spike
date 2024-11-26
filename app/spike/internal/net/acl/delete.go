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
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

func DeletePolicy(source *workloadapi.X509Source, id string) error {
	r := reqres.PolicyDeleteRequest{
		Id: id,
	}

	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New("deletePolicy: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return err
	}

	body, err := net.Post(client, api.UrlPolicyDelete(), mr)
	if err != nil {
		return err
	}

	res := reqres.PolicyDeleteResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return errors.Join(
			errors.New("deletePolicy: Problem parsing response body"),
			err,
		)
	}
	if res.Err != "" {
		return errors.New(string(res.Err))
	}

	return nil
}
