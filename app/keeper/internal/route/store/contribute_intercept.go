//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// guardShardPutRequest validates that the peer contributing a shard is either
// SPIKE Bootstrap or SPIKE Nexus. This prevents unauthorized modification of
// shard data stored in SPIKE Keeper.
//
// Both SPIKE Bootstrap (during initial setup) and SPIKE Nexus (during periodic
// updates) are authorized to contribute shards to SPIKE Keeper.
//
// Parameters:
//   - _ reqres.ShardPutRequest: The request (unused for validation)
//   - w http.ResponseWriter: Response writer for error responses
//   - r *http.Request: The HTTP request containing peer SPIFFE ID
//
// Returns:
//   - error: apiErr.ErrUnauthorized if validation fails, nil otherwise
func guardShardPutRequest(
	_ reqres.ShardPutRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardPutResponse](
		r, w, reqres.ShardPutResponse{
			Err: data.ErrUnauthorized,
		})
	if err != nil {
		return err
	}

	// Allow both Bootstrap (initial setup) and Nexus (periodic updates)
	if !spiffeid.IsBootstrap(peerSPIFFEID.String()) &&
		!spiffeid.IsNexus(peerSPIFFEID.String()) {
		responseBody := net.MarshalBody(reqres.ShardPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
