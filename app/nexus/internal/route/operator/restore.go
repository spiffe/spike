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
)

func RouteRestore(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeRecover"

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

	// TODO: guardRestoreRequest

	shards := recovery.PilotRecoveryShards()

	// TODO: 2 is a magic number; this should be configurable.

	if len(shards) < 2 {
		return errors.ErrNotFound
	}

	payload := shards[:2]

	// TODO: enhancement idea:
	// wait for an acknowledgement from SPIKE Pilot
	// if you get it, either update the database or set up
	// a tombstone indicating that we won't send shards anymore.
	// this way nexus will send the recovery shards only once
	// regardless of who asks them. that's similar to Hashi Vault
	// displaying recovery keys only once during bootstrap.

	responseBody := net.MarshalBody(reqres.RestoreResponse{
		Shards: payload,
	}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "msg", data.ErrSuccess)
	return nil
}
