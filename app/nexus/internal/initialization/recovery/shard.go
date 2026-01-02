//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
)

// shardGetResponse retrieves a shard from a SPIKE Keeper via mTLS POST request.
// It creates an mTLS client using the provided X509 source with a predicate
// that only allows communication with SPIKE Keeper instances.
//
// Parameters:
//   - source: X509Source for mTLS authentication with the keeper
//   - u: The URL of the keeper's shard retrieval endpoint
//
// Returns:
//   - []byte: The raw shard response data from the keeper
//   - *sdkErrors.SDKError: An error if the request fails, nil on success
//
// The function will return an error if:
//   - The X509 source is nil
//   - The request marshaling fails
//   - The POST request fails
//   - The response is empty
func shardGetResponse(
	source *workloadapi.X509Source, u string,
) ([]byte, *sdkErrors.SDKError) {
	if source == nil {
		failErr := sdkErrors.ErrSPIFFENilX509Source.Clone()
		failErr.Msg = "X509 source is nil"
		return nil, failErr
	}

	shardRequest := reqres.ShardGetRequest{}
	md, err := json.Marshal(shardRequest)
	if err != nil {
		failErr := sdkErrors.ErrDataMarshalFailure.Wrap(err)
		failErr.Msg = "failed to marshal shard request"
		return nil, failErr
	}

	client := net.CreateMTLSClientWithPredicate(
		source,
		// Security: Only get shards from SPIKE Keepers.
		predicate.AllowKeeper,
	)

	data, postErr := net.Post(client, u, md)
	if postErr != nil {
		return nil, postErr
	}

	if len(data) == 0 {
		failErr := *sdkErrors.ErrAPIEmptyPayload.Clone()
		failErr.Msg = "received empty shard data from keeper"
		return nil, &failErr
	}

	return data, nil
}
