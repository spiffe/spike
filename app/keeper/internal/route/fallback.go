//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"io"
	"net/http"

	"github.com/spiffe/spike/internal/log"
)

func routeFallback(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeFallback",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	w.WriteHeader(http.StatusBadRequest)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())

	}
}
