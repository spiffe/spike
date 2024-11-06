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

func routeDeleteSecret(r *http.Request, w http.ResponseWriter) {
	fmt.Println("routeDeleteSecret:", r.Method, r.URL.Path, r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretDeleteRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeDeleteSecret: Problem handling request:", err.Error())
		return
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.DeleteSecret(path, versions)
	log.Println("routeDeleteSecret: Secret deleted")

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Println("routeDeleteSecret: Problem writing response:", err.Error())
	}
}
