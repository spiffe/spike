package route

//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func newSecretPutRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.SecretPutRequest {
	var request reqres.SecretPutRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newSecretPutRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routePutSecret(w http.ResponseWriter, r *http.Request) {
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

	req := newSecretPutRequest(requestBody, w)
	if req == nil {
		return
	}

	values := req.Values
	path := req.Path

	state.UpsertSecret(path, values)

	log.Log().Info("routePutSecret", "msg", "Secret upserted")

	res := reqres.SecretPutResponse{}
	responseBody := net.MarshalBody(res, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routePutSecret", "msg", "OK")
}
