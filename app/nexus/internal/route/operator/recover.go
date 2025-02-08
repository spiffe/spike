//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"net/http"
	"sync"
)

var shards []string
var shardsMutex sync.RWMutex

func RouteRecover(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeRestore"

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	// TODO: RecoverResponse should contain # of saved shards
	// and whether recovery was successful.
	// if recovery is not successful it shall reset all shards.
	//

	request := net.HandleRequest[
		reqres.RecoverRequest, reqres.RecoverResponse](
		requestBody, w,
		reqres.RecoverResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	shard := request.Shard

	shardsMutex.Lock()
	shards = append(shards, shard)
	shardsMutex.Unlock()

	if len(shards) == 2 {
		recovery.RecoverBackingStoreUsingPilotShards(shards)
	}

	return nil
}
