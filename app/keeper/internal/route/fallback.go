//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"io"
	"net/http"

	"github.com/spiffe/spike/internal/log"
)

func routeFallback(_ *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())

	}
}
