//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func newCheckInitStateRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.CheckInitStateRequest {
	var request reqres.CheckInitStateRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newCheckInitStateRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.CheckInitStateResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeInitCheck(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInitCheck",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	responseBody := net.ReadRequestBody(r, w)
	if responseBody == nil {
		return
	}

	adminToken := state.AdminToken()

	if adminToken != "" {
		log.Log().Info("routeInitCheck",
			"msg", "Already initialized")

		res := reqres.CheckInitStateResponse{
			State: data.AlreadyInitialized,
		}

		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusOK, responseBody, w)

		log.Log().Info("routeInitCheck",
			"already_initialized", true,
			"msg", "OK",
		)

		return
	}

	res := reqres.CheckInitStateResponse{
		State: data.NotInitialized,
	}

	responseBody = net.MarshalBody(res, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)

	log.Log().Info("routeInitCheck", "msg", "OK")
}
