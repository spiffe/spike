//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func newRootKeyCacheRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.RootKeyCacheRequest {
	var request reqres.RootKeyCacheRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newRootKeyCacheRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.RootKeyCacheResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeKeep(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeKeep",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.RootKeyCacheRequest, reqres.RootKeyCacheResponse](
		requestBody, w,
		reqres.RootKeyCacheResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	rootKey := request.RootKey
	state.SetRootKey(rootKey)

	responseBody := net.MarshalBody(reqres.RootKeyCacheResponse{}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)

	log.Log().Info("routeKeep", "msg", "OK")
}
