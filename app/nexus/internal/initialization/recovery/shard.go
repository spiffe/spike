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
	// TODO: to separate file.
	u, err := url.JoinPath(keeperAPIRoot, string(apiUrl.KeeperShard))
	if err != nil {
		log.Log().Warn(
			fName, "message", "Failed to join path", "url", keeperAPIRoot,
		)
		return ""
	}
	return u
}

func ShardGetResponse(source *workloadapi.X509Source, u string) []byte {
	const fName = "ShardGetResponse"

	if source == nil {
		log.Log().Warn(fName, "message", "Source is nil")
		return []byte{}
	}

	shardRequest := reqres.ShardGetRequest{}
	md, err := json.Marshal(shardRequest)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to marshal request",
			"err", err)
		return []byte{}
	}

	client, err := network.CreateMTLSClientWithPredicate(
		source,
		// Security: Only get shards from SPIKE Keepers.
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

func unmarshalShardResponse(data []byte) *reqres.ShardGetResponse {
	const fName = "unmarshalShardResponse"

	var res reqres.ShardGetResponse
	err := json.Unmarshal(data, &res)
	if err != nil {
		log.Log().Info(fName, "message",
			"Failed to unmarshal response", "err", err)
		return nil
	}
	return &res
}
