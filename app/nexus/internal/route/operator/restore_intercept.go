//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/auth"

	"github.com/spiffe/spike/internal/net"
)

func guardRestoreRequest(
	request reqres.RestoreRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetResponse{
			Err: data.ErrUnauthorized,
		})
	if err != nil {
		return err
	}

	if !spiffeid.IsPilotRestore(peerSPIFFEID.String()) {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.RestoreResponse{
				Err: data.ErrUnauthorized,
			}, w)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	// It's unlikely to have 1000 SPIKE Keepers across the board.
	// The indexes start from 1 and increase one-by-one by design.
	const maxShardID = 1000 // TODO: to constants.
	if request.ID < 1 || request.ID > maxShardID {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.RestoreResponse{
				Err: data.ErrBadInput,
			}, w)
		if err == nil {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	allZero := true
	for _, b := range request.Shard {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.RestoreResponse{
				Err: data.ErrBadInput,
			}, w)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	return nil
}
