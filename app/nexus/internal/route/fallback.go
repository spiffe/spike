//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
)

func routeFallback(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeFallback",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	w.WriteHeader(http.StatusBadRequest)

	res := reqres.FallbackResponse{Err: reqres.ErrBadInput}
	body, err := json.Marshal(res)
	if err != nil {
		res.Err = reqres.ErrServerFault

		log.Log().Error("routeFallback",
			"msg", "Problem generating response",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = io.WriteString(w, string(body))
	if err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())
	}}
}
