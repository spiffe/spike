//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

func routeInit(r *http.Request, w http.ResponseWriter) {
	fmt.Println("routeInit:", r.Method, r.URL.Path, r.URL.RawQuery)

	w.WriteHeader(http.StatusOK)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.AdminTokenWriteRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeInit: Problem handling request:", err.Error())
		return
	}

	adminToken := req.Data          // admin token will be auto created, we just need a strong password, and sanitize that password
	state.SetAdminToken(adminToken) // This is temporary, for demo. Update it based on new sequence diagrams.
	log.Println("routeInit: Admin token saved")

	_, err := io.WriteString(w, "")
	if err != nil {
		log.Println("routeInit: Problem writing response:", err.Error())
	}
}
