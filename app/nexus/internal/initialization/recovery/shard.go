//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"github.com/spiffe/spike-sdk-go/crypto"
	"net/url"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/nexus/internal/env"
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
		func(peerId string) bool {
			return spiffeid.IsKeeper(env.TrustRootForKeeper(), peerId)
		},
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

func shardContributionResponse(
	u string, contribution *[]byte, source *workloadapi.X509Source,
) []byte {
	const fName = "shardContributionResponse"

	client, err := network.CreateMTLSClientWithPredicate(source,
		func(peerId string) bool {
			return spiffeid.IsKeeper(env.TrustRootForKeeper(), peerId)
		},
	)

	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to create mTLS client",
			"err", err)
		return []byte{}
	}

	zeroed := true
	for i := range *contribution {
		if (*contribution)[i] != 0 {
			zeroed = false
			break
		}
	}

	if zeroed {
		log.Log().Info(fName, "message", "All zeros")
		return []byte{}
	}

	if len(*contribution) != crypto.AES256KeySize {
		log.Log().Warn(fName,
			"message", "invalid contribution length",
			"len", len(*contribution))

		// Do not reset `contribution` as this function does not "own" it.

		return []byte{}
	}

	var c [crypto.AES256KeySize]byte
	copy(c[:], *contribution)
	// Security: Ensure that the temporary variable is zeroed out before
	// the function exits.
	defer func() {
		mem.ClearRawBytes(&c)
	}()

	scr := reqres.ShardContributionRequest{
		Shard: &c,
	}
	// Security: Ensure that the struct field is zeroed out before the function
	// exits.
	defer func() {
		mem.ClearRawBytes(scr.Shard)
	}()

	md, err := json.Marshal(scr)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to marshal request",
			"err", err)
		return []byte{}
	}
	// Security: Ensure that the md is zeroed out before the function exits.
	defer func() {
		mem.ClearBytes(md)
	}()

	data, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn(fName, "message",
			"Failed to post",
			"err", err)
	}

	if len(data) == 0 {
		log.Log().Info(fName, "message", "No data")
		return []byte{}
	}

	return data
}
