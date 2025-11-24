//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"

	"github.com/spiffe/spike/internal/net"
)

func ShardGetResponse(
	source *workloadapi.X509Source, u string,
) ([]byte, *sdkErrors.SDKError) {
	if source == nil {
		failErr := sdkErrors.ErrSPIFFENilX509Source
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

	client, err := network.CreateMTLSClientWithPredicate(
		source,
		// Security: Only get shards from SPIKE Keepers.
		predicate.AllowKeeper,
	)
	if err != nil {
		failErr := sdkErrors.ErrNetClientCreationFailed.Wrap(err)
		failErr.Msg = "failed to create mTLS client"
		return nil, failErr
	}

	data, err := net.Post(client, u, md)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		failErr := sdkErrors.ErrDataEmpty
		failErr.Msg = "received empty shard data from keeper"
		return nil, failErr
	}

	return data, nil
}
