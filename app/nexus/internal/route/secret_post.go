package route

//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

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

func routePostSecret(r *http.Request, w http.ResponseWriter) {
	fmt.Println("routePostSecret:", r.Method, r.URL.Path, r.URL.RawQuery)

	w.WriteHeader(http.StatusOK)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretPutRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeInit: Problem handling request:", err.Error())
		return
	}

	values := req.Values
	path := req.Path

	state.UpsertSecret(path, values)
	log.Println("routePostSecret: Secret upserted")

	_, err := io.WriteString(w, "")
	if err != nil {
		log.Println("routePostSecret: Problem writing response:", err.Error())
	}
}
