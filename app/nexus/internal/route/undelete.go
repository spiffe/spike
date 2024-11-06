//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

func routeUndeleteSecret(r *http.Request, w http.ResponseWriter) {
	log.Println("routeUndeleteSecret:", r.Method, r.URL.Path, r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretUndeleteRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeUndeleteSecret: Problem handling request:", err.Error())
		return
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.UndeleteSecret(path, versions)
	log.Println("routeUndeleteSecret: Secret deleted")

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Println("routeUndeleteSecret: Problem writing response:", err.Error())
	}
}
