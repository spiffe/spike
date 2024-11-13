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

	// Start with the default response.
	res := reqres.RootKeyCacheResponse{}
	statusCode := http.StatusOK

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.RootKeyCacheRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		res.Err = reqres.ErrBadInput
		statusCode = http.StatusBadRequest

		body, err := json.Marshal(res)
		if err != nil {
			log.Log().Error("routeKeep",
				"msg", "Problem generating response",
				"err", err.Error())

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(`{"error":"internal server error"}`))
			if err != nil {
				log.Log().Error("routeKeep",
					"msg", "Problem writing response",
					"err", err.Error())
				return
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeKeep",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return
	}

	rootKey := req.RootKey
	state.SetRootKey(rootKey)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)

	statusCode = http.StatusOK
	res.Err = ""

	body, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem generating response",
			"err", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(`{"error":"internal server error"}`))
		if err != nil {
			log.Log().Error("routeKeep",
				"msg", "Problem writing response",
				"err", err.Error())
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	statusCode = http.StatusOK
	res.Err = ""

	_, err = io.WriteString(w, string(body))
	if err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem writing response:",
			"err", err.Error())
		return
	}
	log.Log().Info("routeKeep", "msg", "OK")
}
