//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/auth"
)

// guardShardGetRequest validates that the peer requesting shard retrieval is
// SPIKE Nexus. This prevents unauthorized access to sensitive shard data
// stored in SPIKE Keeper.
//
// Only SPIKE Nexus is authorized to retrieve shards from SPIKE Keeper during
// recovery operations.
//
// Parameters:
//   - _ reqres.ShardGetRequest: The request (unused for validation)
//   - w http.ResponseWriter: Response writer for error responses
//   - r *http.Request: The HTTP request containing peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: sdkErrors.ErrAccessUnauthorized if validation fails,
//     nil otherwise
func guardShardGetRequest(
	_ reqres.ShardGetRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Only SPIKE Nexus is authorized to retrieve shards from SPIKE Keeper.
	if !spiffeid.IsNexus(peerSPIFFEID.String()) {
		failErr := net.Fail(
			reqres.ShardGetResponse{}.Unauthorized(), w, http.StatusUnauthorized,
		)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized.Clone()
	}

	return nil
}
