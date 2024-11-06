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

func routeListPaths(r *http.Request, w http.ResponseWriter) {
	log.Println("routeListPaths:", r.Method, r.URL.Path, r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretListRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeListPaths: Problem handling request:", err.Error())
		return
	}

	keys := state.ListKeys()

	res := reqres.SecretListResponse{Keys: keys}
	md, err := json.Marshal(res)
	if err != nil {
		log.Println("routeListPaths: Problem generating response:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Println("routeListPaths: Problem writing response:", err.Error())
	}
}
