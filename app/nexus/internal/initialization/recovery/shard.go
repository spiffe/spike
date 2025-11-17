//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"

	"github.com/spiffe/spike/internal/net"
)

func ShardGetResponse(source *workloadapi.X509Source, u string) []byte {
	const fName = "ShardGetResponse"

	if source == nil {
		log.Log().Warn(fName, "message", "source is nil")
		return []byte{}
	}

	shardRequest := reqres.ShardGetRequest{}
	md, err := json.Marshal(shardRequest)
	if err != nil {
		log.Log().Warn(fName,
			"message", "failed to marshal request",
			"err", err,
		)
		return []byte{}
	}

	client, err := network.CreateMTLSClientWithPredicate(
		source,
		// Security: Only get shards from SPIKE Keepers.
		predicate.AllowKeeper,
	)
	if err != nil {
		log.Log().Warn(
			fName,
			"message", "failed to create mTLS client",
			"err", err,
		)
		return []byte{}
	}

	data, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn(
			fName,
			"message", "failed to post",
			"err", err,
		)
	}

	if len(data) == 0 {
		log.Log().Info(fName, "message", "mo data")
		return []byte{}
	}

	return data
}
