//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
)

// unmarshalShardResponse deserializes JSON data into a ShardGetResponse
// structure.
//
// This function is used during the recovery process to parse HTTP responses
// from SPIKE Keeper instances when retrieving Shamir secret shards. If the
// JSON unmarshaling fails, the error is logged and nil is returned.
//
// Parameters:
//   - data: The raw JSON response body from a keeper shard endpoint
//
// Returns:
//   - *reqres.ShardGetResponse: A pointer to the deserialized response
//     containing the shard data, or nil if unmarshaling fails
func unmarshalShardResponse(data []byte) *reqres.ShardGetResponse {
	const fName = "unmarshalShardResponse"

	var res reqres.ShardGetResponse

	err := json.Unmarshal(data, &res)
	if err != nil {
		log.Log().Info(fName, "message", "failed to unmarshal response", "err", err)
		return nil
	}

	return &res
}
