//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func newSecretDeleteRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.SecretDeleteRequest {
	var request reqres.SecretDeleteRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newSecretDeleteRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.SecretDeleteResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeDeleteSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeDeleteSecret",
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

	request := newSecretDeleteRequest(requestBody, w)

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.DeleteSecret(path, versions)
	log.Log().Info("routeDeleteSecret",
		"msg", "Secret deleted")

	responseBody := net.MarshalBody(reqres.SecretDeleteResponse{}, w)

	net.Respond(http.StatusOK, responseBody, w)

	log.Log().Info("routeDeleteSecret", "msg", "OK")
}
