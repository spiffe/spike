//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

//func Contribute(source *workloadapi.X509Source) {
//	peers := env.Peers()
//	myId := env.KeeperId()
//
//	for id, peer := range peers {
//		if id == myId {
//			continue
//		}
//
//		contributeUrl, err := url.JoinPath(peer, "v1/store/contribute")
//		if err != nil {
//			log.FatalLn("Failed to join path: " + err.Error())
//		}
//
//		if source == nil {
//			log.FatalLn("contribute: source is nil")
//		}
//
//		client, err := net.CreateMtlsClientWithPredicate(
//			source,
//			auth.IsKeeper,
//		)
//		if err != nil {
//			panic(err)
//		}
//
//		contribution := state.RandomContribution()
//		state.Shards.Store(myId, contribution)
//
//		log.Log().Info(
//			"contribute",
//			"msg", "Sending contribution to peer",
//			"peer", peer,
//		)
//
//		md, err := json.Marshal(
//			reqres.ShardContributionRequest{
//				KeeperId: myId,
//				Shard:    base64.StdEncoding.EncodeToString(contribution),
//			},
//		)
//
//		_, err = net.Post(client, contributeUrl, md)
//		for err != nil {
//			time.Sleep(5 * time.Second)
//			_, err = net.Post(client, contributeUrl, md)
//			if err != nil {
//				log.Log().Info("contribute",
//					"msg", "Error sending contribution. Will retry",
//					"err", err,
//				)
//				time.Sleep(5 * time.Second)
//			}
//		}
//	}
//}
