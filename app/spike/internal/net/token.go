//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// SendInitRequest sends an init request to SPIKE Nexus.
func SendInitRequest(source *workloadapi.X509Source, token string) error {
	//TODO: log.Println("##### THE SIGNATURE OF THE INIT REQUEST WILL CHANGE ####")
	r := reqres.AdminTokenWriteRequest{
		Data: token,
	}
	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New("token: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, config.CanTalkToPilot)

	_, err = net.Post(client, urlInit, mr)
	return err
}
