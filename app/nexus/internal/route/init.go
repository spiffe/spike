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

func routeInit(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInit",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	w.WriteHeader(http.StatusOK)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.AdminTokenWriteRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Info("routeInit",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	adminToken := req.Data          // admin token will be auto created, we just need a strong password, and sanitize that password
	state.SetAdminToken(adminToken) // This is temporary, for demo. Update it based on new sequence diagrams.
	log.Log().Info("routeInit", "msg", "Admin token saved")

	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routeInit",
			"msg", "Problem writing response", "err", err.Error())
	}
}
