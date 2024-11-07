//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeKeep(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeKeep",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.RootKeyCacheRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeKeep",
				"msg", "Problem writing response",
				"err", err.Error())
		}
		return
	}

	rootKey := req.RootKey
	state.SetRootKey(rootKey)

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "OK")
	if err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem writing response:",
			"err", err.Error())
		return
	}
	log.Log().Info("routeKeep", "OK")
}
