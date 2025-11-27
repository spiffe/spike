//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// unmarshalShardResponse deserializes JSON data into a ShardGetResponse
// structure.
//
// This function is used during the recovery process to parse HTTP responses
// from SPIKE Keeper instances when retrieving Shamir secret shards.
//
// Parameters:
//   - data: The raw JSON response body from a keeper shard endpoint
//
// Returns:
//   - *reqres.ShardGetResponse: A pointer to the deserialized response
//     containing the shard data, or nil if unmarshaling fails
//   - *sdkErrors.SDKError: An error if JSON unmarshaling fails, nil on success
func unmarshalShardResponse(data []byte) (
	*reqres.ShardGetResponse, *sdkErrors.SDKError,
) {
	var res reqres.ShardGetResponse

	err := json.Unmarshal(data, &res)
	if err != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		failErr.Msg = "failed to unmarshal response"
		return nil, failErr
	}

	return &res, nil
}
