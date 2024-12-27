//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/base64"
	"encoding/json"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"net/url"
	"time"
)

func Contribute(source *workloadapi.X509Source) {
	peers := env.Peers()
	myId := env.KeeperId()

	for id, peer := range peers {
		if id == myId {
			continue
		}

		// TODO: error handling.
		contributeUrl, _ := url.JoinPath(peer, "v1/store/contribute")

		// TODO: log.Fatalf instead of panicking.
		if source == nil {
			panic("contribute: No source")
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
				// TODO: version is not needed in the new algorithm.
				Version: 0,
			},
		)

		// TODO: this is temporary; we need a more robust handling.
		// maybe re-use our exponential backoff library.
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
