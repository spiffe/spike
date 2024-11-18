//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteInit handles the initial setup of the system, creating admin credentials
// and tokens. This endpoint can only be called once - subsequent calls will
// fail.
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
//   - 500 Internal Server Error: System already initialized
//     (err: "already_initialized")
//   - 500 Internal Server Error: Failed to generate secure random values
//     (err: "server_fault")
//
// The function uses cryptographically secure random number generation for both
// the admin token and salt. The admin token is prefixed with "spike." before
// storage. All operations are logged using structured logging.
func RouteInit(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeInit", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditCreate

	// No need to check for valid JWT here. System initialization is done
	// anonymously by the first user (who will be the admin).
	// If the system is already initialized, this process will err out anyway.

	requestBody, err := prepareInitRequestBody(w, r)
	if err != nil {
		return err
	}

	request, err := prepareInitRequest(requestBody, w)
	if err != nil {
		return err
	}

	request, err = sanitizeInitRequest(request, w)
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

	// Generate salt and hash password
	salt, err := generateSalt(w)
	if err != nil {
		return err
	}

	updateStateForInit(request.Password, adminTokenBytes, salt)

	responseBody := net.MarshalBody(reqres.InitResponse{}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInit", "msg", "OK")
	return nil
}
