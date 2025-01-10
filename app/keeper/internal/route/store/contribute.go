//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/base64"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteContribute(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeContribute"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.ShardContributionRequest, reqres.ShardContributionResponse](
		requestBody, w,
		reqres.ShardContributionResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	shard := request.Shard
	id := request.KeeperId

	// TODO: maybe sanitize these.
	// keeper id should be an integer
	// shard should have a certain size, etc.

	// Decode shard content from Base64 encoding.
	decodedShard, err := base64.StdEncoding.DecodeString(shard)
	if err != nil {
		log.Log().Error(fName, "msg", "Failed to decode shard", "err", err.Error())
		http.Error(w, "Invalid shard content", http.StatusBadRequest)
		return errors.ErrParseFailure
	}

	// Store decoded shard in the map.
	state.Shards.Store(id, decodedShard)
	log.Log().Info(fName, "msg", "Shard stored", "id", id)

	responseBody := net.MarshalBody(reqres.ShardContributionResponse{}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)

	return nil
}
