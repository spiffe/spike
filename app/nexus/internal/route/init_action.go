//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"golang.org/x/crypto/pbkdf2"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routeInit handles the initial setup of the system, creating admin credentials
// and tokens. This endpoint can only be called once - subsequent calls will fail.
//
// The function performs system initialization by:
//  1. Validating the provided admin password meets security requirements
//  2. Generating a secure random admin token
//  3. Creating a password hash using PBKDF2-SHA256 with secure parameters
//  4. Storing the credentials in the system state
//
// Security parameters:
//   - Minimum password length: 16 characters
//   - Admin token length: 32 bytes (256 bits)
//   - Salt length: 16 bytes (128 bits)
//   - PBKDF2 iterations: 600,000 (OWASP minimum recommendation)
//   - Hash output length: 32 bytes (256 bits)
//
// Request body format:
//
//	{
//	    "password": string   // Admin password (min 16 chars)
//	}
//
// Response format on success (200 OK):
//
//	{} // Empty response indicates success
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body
//   - 400 Bad Request: Password too short (err: "low_entropy")
//   - 500 Internal Server Error: System already initialized (err: "already_initialized")
//   - 500 Internal Server Error: Failed to generate secure random values (err: "server_fault")
//
// The function uses cryptographically secure random number generation for both
// the admin token and salt. The admin token is prefixed with "spike." before storage.
// All operations are logged using structured logging.
func routeInit(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInit", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	// No need to check for valid JWT here. System initialization is done
	// anonymously by the first user (who will be the admin).
	// If the system is already initialized, this process will err out anyway.

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.InitRequest, reqres.InitResponse](
		requestBody, w,
		reqres.InitResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	password := request.Password
	if len(password) < 16 {
		res := reqres.InitResponse{Err: reqres.ErrLowEntropy}

		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Password too short")
		return
	}

	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInit", "msg", "Already initialized")

		responseBody := net.MarshalBody(
			reqres.InitResponse{Err: reqres.ErrAlreadyInitialized}, w,
		)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Already initialized")
		return
	}

	log.Log().Info("routeInit", "msg", "No admin token. will create one")

	// Generate adminToken (32 bytes)
	adminTokenBytes := make([]byte, 32)
	if _, err := rand.Read(adminTokenBytes); err != nil {
		log.Log().Error("routeInit",
			"msg", "Failed to generate admin token", "err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: reqres.ErrServerFault}, w,
		)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Failed to generate admin token")
		return
	}

	// Generate salt and hash password
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Log().Error("routeInit", "msg", "Failed to generate salt",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: reqres.ErrServerFault}, w,
		)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Failed to generate salt")
		return
	}

	// TODO: make this configurable.
	iterationCount := 600_000 // Minimum OWASP recommendation for PBKDF2-SHA256
	hashLength := 32          // 256 bits output

	passwordHash := pbkdf2.Key(
		[]byte(password), salt,
		iterationCount, hashLength, sha256.New,
	)

	state.SetAdminToken("spike." + string(adminTokenBytes))
	state.SetAdminCredentials(
		hex.EncodeToString(passwordHash),
		hex.EncodeToString(salt),
	)

	responseBody := net.MarshalBody(reqres.InitResponse{}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInit", "msg", "OK")
}
