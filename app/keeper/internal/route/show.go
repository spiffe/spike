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

func newRootKeyReadRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.RootKeyReadRequest {
	var request reqres.RootKeyReadRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newRootKeyReadRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.RootKeyReadResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeShow(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeShow",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := newRootKeyReadRequest(requestBody, w)
	if request == nil {
		return
	}

	rootKey := state.RootKey()

	responseBody := net.MarshalBody(
		reqres.RootKeyReadResponse{RootKey: rootKey}, w,
	)

	net.Respond(http.StatusOK, responseBody, w)

	log.Log().Info("routeShow", "msg", "OK")
}
