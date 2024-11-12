package route

//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routePutSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeGetSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := ensureValidJwt(w, r)
	if !validJwt {
		return
	}

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretPutRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routePutSecret",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	values := req.Values
	path := req.Path

	state.UpsertSecret(path, values)

	log.Log().Info("routePutSecret", "msg", "Secret upserted")

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routePutSecret",
			"msg", "Problem writing response", "err", err.Error())
	}
}
