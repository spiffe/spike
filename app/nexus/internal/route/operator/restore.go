//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

var shards []string
var shardsMutex sync.RWMutex

func RouteRestore(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeRestore"

	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.RestoreRequest, reqres.RestoreResponse](
		requestBody, w,
		reqres.RestoreResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	shard := request.Shard

	shardsMutex.Lock()
	shards = append(shards, shard)
	shardsMutex.Unlock()

	if len(shards) == 2 {
		recovery.RestoreBackingStoreUsingPilotShards(shards)
	}

	responseBody := net.MarshalBody(reqres.RestoreResponse{
		RestorationStatus: data.RestorationStatus{
			ShardsCollected: len(shards),
			ShardsRemaining: 2 - len(shards),
			Restored:        len(shards) == 2,
		},
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
