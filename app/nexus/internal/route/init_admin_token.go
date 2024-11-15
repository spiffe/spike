//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/rand"
	"errors"
	"github.com/spiffe/spike/app/nexus/internal/config"
	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

func checkAdminToken(w http.ResponseWriter) error {
	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInit", "msg", "Already initialized")

		responseBody := net.MarshalBody(
			reqres.InitResponse{Err: reqres.ErrAlreadyInitialized}, w,
		)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Already initialized")
		return errors.New("already initialized")
	}
	return nil
}

func generateAdminToken(w http.ResponseWriter) ([]byte, error) {
	// Generate adminToken (32 bytes)
	adminTokenBytes := make([]byte, config.SpikeNexusAdminTokenBytes)
	if _, err := rand.Read(adminTokenBytes); err != nil {
		log.Log().Error("routeInit",
			"msg", "Failed to generate admin token", "err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: reqres.ErrServerFault}, w,
		)
		if responseBody == nil {
			return []byte{}, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info(
			"routeInit", "msg", "exit: Failed to generate admin token",
		)
		return []byte{}, errors.New("failed to generate admin token")
	}
	return adminTokenBytes, nil
}
