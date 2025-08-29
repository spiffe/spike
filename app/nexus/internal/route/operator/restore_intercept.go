//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/net"
)

func guardRestoreRequest(
	request reqres.RestoreRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if peerSPIFFEID == nil {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if !spiffeid.IsPilotRestore(env.TrustRootForPilot(), peerSPIFFEID.String()) {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	// It's unlikely to have 1000 SPIKE Keepers across the board.
	// The indexes start from 1 and increase one-by-one by design.
	const maxShardID = 1000

	if request.ID < 1 || request.ID > maxShardID {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
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
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	return nil
}
