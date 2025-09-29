//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"

	"github.com/spiffe/spike/internal/net"
)

func shardURL(keeperAPIRoot string) string {
	const fName = "shardURL"

	u, err := url.JoinPath(keeperAPIRoot, string(apiUrl.KeeperShard))
	if err != nil {
		log.Log().Warn(
			fName, "message", "Failed to join path", "url", keeperAPIRoot,
		)
		return ""
	}
	return u
}

func shardResponse(source *workloadapi.X509Source, u string) []byte {
	const fName = "shardResponse"

	shardRequest := reqres.ShardRequest{}
	md, err := json.Marshal(shardRequest)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to marshal request",
			"err", err)
		return []byte{}
	}

	client, err := network.CreateMTLSClientWithPredicate(
		source,
		// Only get shards from SPIKE Keepers.
		predicate.AllowKeeper,
	)

	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to create mTLS client",
			"err", err)
		return []byte{}
	}

	data, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to post",
			"err", err)
	}

	if len(data) == 0 {
		log.Log().Info(fName, "message", "No data")
		return []byte{}
	}

	return data
}

func unmarshalShardResponse(data []byte) *reqres.ShardResponse {
	const fName = "unmarshalShardResponse"

	var res reqres.ShardResponse
	err := json.Unmarshal(data, &res)
	if err != nil {
		log.Log().Info(fName, "message",
			"Failed to unmarshal response", "err", err)
		return nil
	}
	return &res
}
