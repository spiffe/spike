//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeGetSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeGetSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretReadRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	version := req.Version
	path := req.Path

	secret, exists := state.GetSecret(path, version)
	if !exists {
		log.Log().Info("routeGetSecret", "msg", "Secret not found")
		w.WriteHeader(http.StatusNotFound)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	res := reqres.SecretReadResponse{Data: secret}
	md, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem generating response", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Log().Info("routeGetSecret", "msg", "Got secret")

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem writing response", "err", err.Error())
	}
}
