//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"golang.org/x/crypto/pbkdf2"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func prepareRequestBody(
	w http.ResponseWriter, r *http.Request,
) ([]byte, error) {
	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return []byte{}, errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.InitRequest, reqres.InitResponse](
		requestBody, w,
		reqres.InitResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return []byte{}, errors.New("failed to parse request body")
	}
	return requestBody, nil
}

func prepareRequest(
	requestBody []byte, w http.ResponseWriter,
) (*reqres.InitRequest, error) {
	request := net.HandleRequest[
		reqres.InitRequest, reqres.InitResponse](
		requestBody, w,
		reqres.InitResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return nil, errors.New("failed to parse request body")
	}
	return request, nil
}

func sanitizeRequest(
	req *reqres.InitRequest, w http.ResponseWriter,
) (*reqres.InitRequest, error) {
	password := req.Password
	if len(password) < 16 { // TODO: magic number
		res := reqres.InitResponse{Err: reqres.ErrLowEntropy}

		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return nil, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Password too short")
		return nil, errors.New("password too short")
	}

	return req, nil
}

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
	adminTokenBytes := make([]byte, 32) // TODO: magic number.
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
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
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
func routeInit(w http.ResponseWriter, r *http.Request, audit *log.AuditEntry) error {
	log.Log().Info("routeInit", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = "create"

	// No need to check for valid JWT here. System initialization is done
	// anonymously by the first user (who will be the admin).
	// If the system is already initialized, this process will err out anyway.

	requestBody, err := prepareRequestBody(w, r)
	if err != nil {
		return err
	}

	request, err := prepareRequest(requestBody, w)
	if err != nil {
		return err
	}

	request, err = sanitizeRequest(request, w)
	if err != nil {
		return err
	}

	err = checkAdminToken(w)
	if err != nil {
		return err
	}

	log.Log().Info("routeInit", "msg", "No admin token. will create one")

	adminTokenBytes, err := generateAdminToken(w)
	if err != nil {
		return err
	}

	>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> I WAS LEFT HERE!!!!!

	// Generate salt and hash password
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Log().Error("routeInit", "msg", "Failed to generate salt",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: reqres.ErrServerFault}, w,
		)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Failed to generate salt")
		return errors.New("failed to generate salt")
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
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInit", "msg", "OK")
	return nil
}
