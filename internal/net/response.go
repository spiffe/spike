//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"net/http"
)

func MarshalBody(res any, w http.ResponseWriter) []byte {
	body, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("marshalBody",
			"msg", "Problem generating response",
			"err", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(`{"error":"internal server error"}`))
		if err != nil {
			log.Log().Error("marshalBody",
				"msg", "Problem writing response",
				"err", err.Error())
			return nil
		}
		return nil
	}
	return body
}

func Respond(statusCode int, body []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, err := w.Write(body)
	if err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}

func Fallback(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("fallback",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := MarshalBody(reqres.FallbackResponse{Err: reqres.ErrBadInput}, w)
	if body == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if _, err := w.Write(body); err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}
