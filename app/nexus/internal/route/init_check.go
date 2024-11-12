//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"github.com/spiffe/spike/internal/entity/data"
	"io"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeInitCheck(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInitCheck",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.CheckInitStateRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Info("routeInitCheck",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	adminToken := state.AdminToken()

	if adminToken != "" {
		log.Log().Info("routeInitCheck",
			"msg", "Already initialized")

		res := reqres.CheckInitStateResponse{
			State: data.AlreadyInitialized,
		}
		md, err := json.Marshal(res)
		if err != nil {
			res.Err = reqres.ErrServerFault

			log.Log().Error("routeInitCheck",
				"msg", "Problem generating response", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		_, err = io.WriteString(w, string(md))
		if err != nil {
			log.Log().Error("routeInitCheck",
				"msg", "Problem writing response", "err", err.Error())
		}

		return
	}

	res := reqres.CheckInitStateResponse{
		State: data.NotInitialized,
	}
	md, err := json.Marshal(res)

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeInitCheck",
			"msg", "Problem writing response", "err", err.Error())
	}

}
