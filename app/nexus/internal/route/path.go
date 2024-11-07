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

func routeListPaths(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeListPaths",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretListRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeListPaths",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	keys := state.ListKeys()

	res := reqres.SecretListResponse{Keys: keys}
	md, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("routeListPaths",
			"msg", "Problem generating response",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeListPaths",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}
