//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/net"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
)

// Init sends an init request to SPIKE Nexus.
func Init(source *workloadapi.X509Source, password string) error {
	r := reqres.InitRequest{
		Password: password,
	}
	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New("init: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, config.CanTalkToPilot)

	_, err = net.Post(client, urlInit, mr)

	return err
}

func CheckInitState(source *workloadapi.X509Source) (data.InitState, error) {
	r := reqres.CheckInitStateRequest{}
	mr, err := json.Marshal(r)
	if err != nil {
		return data.NotInitialized, errors.Join(
			errors.New("checkInitState: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, config.CanTalkToPilot)

	body, err := net.Post(client, urlInitState, mr)
	if err != nil {
		return data.NotInitialized, errors.Join(
			errors.New("checkInitState: I am having problem sending the request"), err)
	}

	var res reqres.CheckInitStateResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return data.NotInitialized, errors.Join(
			errors.New("checkInitState: Problem parsing response body"),
			err,
		)
	}

	state := res.State

	return state, nil
}
