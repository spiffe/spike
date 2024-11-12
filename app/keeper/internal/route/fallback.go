//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
)

// TODO: there are multiple WriteHeader's in heandlers which will cause issues.
// Follow this example to fix all of them.
// The handlers are in the `route` package.
// TODO: this fallback route is pretty generic; move it to a common place.
func routeFallback(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeFallback",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	// Start with the default response
	res := reqres.FallbackResponse{Err: reqres.ErrBadInput}
	statusCode := http.StatusBadRequest

	// Try to marshal the response
	body, err := json.Marshal(res)
	if err != nil {
		// Update both the response and status code if marshal fails
		res.Err = reqres.ErrServerFault
		statusCode = http.StatusInternalServerError

		log.Log().Error("routeFallback",
			"msg", "Problem generating response",
			"err", err.Error())

		// Try to marshal the error response
		body, err = json.Marshal(res)
		if err != nil {
			// If even error marshaling fails, send a basic error message
			w.WriteHeader(statusCode)
			_, err = w.Write([]byte(`{"error":"internal server error"}`))
			if err != nil {
				log.Log().Error("routeFallback",
					"msg", "Problem writing response",
					"err", err.Error())
				return
			}
			return
		}
	}

	// Set content type header before writing status and body
	w.Header().Set("Content-Type", "application/json")

	// Now we can safely write the status code once
	w.WriteHeader(statusCode)

	// Write the response body
	if _, err := w.Write(body); err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())
		// Note: Can't change status code here as it's already been sent
	}
}
