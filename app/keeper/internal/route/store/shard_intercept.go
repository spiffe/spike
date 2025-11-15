//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
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
//   - error: apiErr.ErrUnauthorized if validation fails, nil otherwise
func guardShardGetRequest(
	_ reqres.ShardGetRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Only SPIKE Nexus is authorized to retrieve shards from SPIKE Keeper.
	if err := net.FailIf(
		!spiffeid.IsNexus(peerSPIFFEID.String()),
		reqres.ShardGetUnauthorized, w,
		http.StatusUnauthorized, apiErr.ErrUnauthorized,
	); err != nil {
		return err
	}

	return nil
}
