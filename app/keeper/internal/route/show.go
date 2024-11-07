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

func routeShow(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeShow",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.RootKeyReadRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeShow",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeShow",
				"msg", "Problem writing response",
				"err", err.Error())
		}
		return
	}

	rootKey := state.RootKey()

	res := reqres.RootKeyReadResponse{RootKey: rootKey}
	md, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("routeShow",
			"msg", "Problem generating response",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeShow",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}
