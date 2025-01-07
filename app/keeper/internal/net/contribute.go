//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

func Contribute(source *workloadapi.X509Source) {
	peers := env.Peers()
	myId := env.KeeperId()

	for id, peer := range peers {
		if id == myId {
			continue
		}

		contributeUrl, err := url.JoinPath(peer, "v1/store/contribute")
		if err != nil {
			log.FatalLn("Failed to join path: " + err.Error())
		}

		if source == nil {
			log.FatalLn("contribute: source is nil")
		}

		client, err := net.CreateMtlsClientWithPredicate(
			source,
			auth.IsKeeper,
		)
		if err != nil {
			panic(err)
		}

		contribution := state.RandomContribution()
		state.Shards.Store(myId, contribution)

		log.Log().Info(
			"contribute",
			"msg", "Sending contribution to peer",
			"peer", peer,
		)

		md, err := json.Marshal(
			reqres.ShardContributionRequest{
				KeeperId: myId,
				Shard:    base64.StdEncoding.EncodeToString(contribution),
			},
		)
		if err != nil {
			log.FatalLn(
				"Failed to marshal shard contribution request: " + err.Error(),
			)
		}

		_, err = net.Post(client, contributeUrl, md)
		for err != nil {
			time.Sleep(5 * time.Second)
			_, err = net.Post(client, contributeUrl, md)
			if err != nil {
				log.Log().Info("contribute",
					"msg", "Error sending contribution. Will retry",
					"err", err,
				)
				time.Sleep(5 * time.Second)
			}
		}
	}
}
