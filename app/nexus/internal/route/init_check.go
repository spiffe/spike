//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeInitCheck(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInitCheck", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	// No need to check for valid JWT here. System initialization is done
	// anonymously by the first user (who will be the admin).
	// If the system is already initialized, this process will err out anyway.

	responseBody := net.ReadRequestBody(r, w)
	if responseBody == nil {
		return
	}

	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInitCheck", "msg", "Already initialized")

		responseBody := net.MarshalBody(reqres.CheckInitStateResponse{
			State: data.AlreadyInitialized}, w,
		)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusOK, responseBody, w)
		log.Log().Info("routeInitCheck",
			"already_initialized", true,
			"msg", "OK",
		)
		return
	}

	responseBody = net.MarshalBody(reqres.CheckInitStateResponse{
		State: data.NotInitialized,
	}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInitCheck", "msg", "OK")
}
