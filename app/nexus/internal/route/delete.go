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

func routeDeleteSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeDeleteSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := ensureValidJwt(w, r)
	if !validJwt {
		w.WriteHeader(http.StatusUnauthorized)

		res := reqres.SecretDeleteResponse{
			Err: reqres.ErrUnauthorized,
		}

		body, err := json.Marshal(res)
		if err != nil {
			res.Err = reqres.ErrServerFault

			log.Log().Error("routeDeleteSecret",
				"msg", "Problem generating response",
				"err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeDeleteSecret",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return
	}

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretDeleteRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeDeleteSecret",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.DeleteSecret(path, versions)
	log.Log().Info("routeDeleteSecret",
		"msg", "Secret deleted")

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routeDeleteSecret",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}
