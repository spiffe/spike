//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spiffe/spike/internal/config"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	api "github.com/spiffe/spike/app/spike/internal/net"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// Init sends an init request to SPIKE Nexus.
func Init(source *workloadapi.X509Source) error {
	r := reqres.InitRequest{}
	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New("init: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.CanTalkToPilot)

	body, err := net.Post(client, api.UrlInit(), mr)

	var res reqres.InitResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return errors.Join(
			errors.New("init: Problem parsing response body"),
			err,
		)
	}

	if res.RecoveryToken == "" {
		fmt.Println("Failed to get recovery token")
		return errors.New("failed to get recovery token")
	}

	err = os.WriteFile(
		config.SpikePilotRootKeyRecoveryFile(), []byte(res.RecoveryToken), 0600,
	)
	if err != nil {
		fmt.Println("Failed to save token to file:")
		fmt.Println(err.Error())
		return errors.New("failed to save token to file")
	}

	return nil
}

// CheckInitState sends a checkInitState request to SPIKE Nexus.
func CheckInitState(source *workloadapi.X509Source) (data.InitState, error) {
	r := reqres.CheckInitStateRequest{}
	mr, err := json.Marshal(r)
	if err != nil {
		return data.NotInitialized, errors.Join(
			errors.New("checkInitState: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.CanTalkToPilot)

	fmt.Println(">>>>>>>>>>>>>>>Posting to", api.UrlInitState())

	body, err := net.Post(client, api.UrlInitState(), mr)
	if errors.Is(err, net.ErrUnauthorized) {
		return data.NotInitialized, err
	}

	if err != nil {
		return data.NotInitialized, errors.Join(
			errors.New(
				"checkInitState: I am having problem sending the request",
			), err)
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
