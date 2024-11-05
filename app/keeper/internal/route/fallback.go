//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"io"
	"log"
	"net/http"
)

func routeFallback(_ *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Println("routeFallback: Problem writing response:", err.Error())
	}
}
