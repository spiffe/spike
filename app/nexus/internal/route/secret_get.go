//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

func newSecretReadRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.SecretReadRequest {
	var request reqres.SecretReadRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newSecretReadRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.SecretReadResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeGetSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeGetSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := newSecretReadRequest(requestBody, w)
	if request == nil {
		return
	}

	version := request.Version
	path := request.Path

	secret, exists := state.GetSecret(path, version)
	if !exists {
		log.Log().Info("routeGetSecret", "msg", "Secret not found")

		res := reqres.SecretReadResponse{Err: reqres.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		log.Log().Info("routeGetSecret", "msg", "not found")

		return
	}

	res := reqres.SecretReadResponse{Data: secret}
	responseBody := net.MarshalBody(res, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetSecret", "msg", "OK")
}
