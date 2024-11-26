//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package init

import (
	"crypto/rand"
	"errors"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/config"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// checkPreviousInitialization verifies if an admin token already exists in the
// system state. If an admin token is present, it responds with an
// "already initialized" error through the provided http.ResponseWriter and
// returns an error.
//
// The function handles the HTTP response writing internally, setting
// appropriate status codes and error messages.
//
// Returns nil if no admin token exists, otherwise returns an error indicating
// the system is already initialized.
func checkPreviousInitialization(w http.ResponseWriter) error {
	if state.Initialized() {
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

// generateAdminSigningToken creates a new random admin token of length specified by
// config.SpikeNexusAdminTokenBytes.
//
// It handles error responses through the provided http.ResponseWriter in case
// of failures during token generation or response marshaling.
//
// Returns:
//   - []byte: The generated admin token bytes
//   - error: nil on success, otherwise an error describing what went wrong
//
// If token generation fails, it will set an appropriate HTTP error response
// and return an empty byte slice along with an error.
func generateAdminSigningToken(w http.ResponseWriter) ([]byte, error) {
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
