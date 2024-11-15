//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func fetchAdminToken(w http.ResponseWriter) (string, error) {
	adminToken := state.AdminToken()
	if adminToken == "" {
		log.Log().Error("routeAdminLogin", "msg", "Admin token not set")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return "", errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return "", errors.New("admin token not set")
	}
	return adminToken, nil
}
