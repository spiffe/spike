//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeUndeleteSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeUndeleteSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	req := net.HandleRequest[
		reqres.SecretUndeleteRequest, reqres.SecretUndeleteResponse](
		requestBody, w,
		reqres.SecretUndeleteResponse{Err: reqres.ErrBadInput},
	)

	req := newSecretUndeleteRequest(requestBody, w)
	if req == nil {
		return
	}

	path := req.Path
	versions := req.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	state.UndeleteSecret(path, versions)
	log.Log().Info("routeUndeleteSecret", "msg", "Secret undeleted")

	res := reqres.SecretUndeleteResponse{}
	responseBody := net.MarshalBody(res, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeUndeleteSecret", "msg", "OK")
}
