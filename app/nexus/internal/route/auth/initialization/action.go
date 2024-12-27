//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"crypto/rand"
	"errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// generateSalt creates a new random 16-byte salt value using crypto/rand.
//
// It handles error responses through the provided http.ResponseWriter in case
// of failures during salt generation or response marshaling.
//
// Returns:
//   - []byte: The generated salt bytes
//   - error: nil on success, otherwise an error describing what went wrong
//
// If salt generation fails, it will set an appropriate HTTP error response
// with a 500 status code and return an empty byte slice along with an error.
func generateSalt(w http.ResponseWriter) ([]byte, error) {
	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		log.Log().Error("routeInit", "msg", "Failed to generate salt",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: data.ErrInternal}, w,
		)
		if responseBody == nil {
			return []byte{}, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Failed to generate salt")
		return []byte{}, errors.New("failed to generate salt")
	}

	return salt, nil
}

// RouteInit handles the initial setup of the system. This endpoint can only be
// called once - subsequent calls will fail.
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
	// TODO: we don't need init anymore.

	return nil

	//// This flow will change after implementing Shamir Secrets Sharing
	//// `init` will ensure there are enough keepers connected, and then
	//// initialize the keeper instances.
	////
	//// We will NOT need the encrypted root key; instead, an admin user will
	//// fetch enough shards to back up. Admin will need to provide some sort
	//// of key or password to get the data in encrypted form.
	//
	//log.Log().Info("routeInit", "method", r.Method, "path", r.URL.Path,
	//	"query", r.URL.RawQuery)
	//audit.Action = log.AuditCreate
	//
	//requestBody, err := prepareInitRequestBody(w, r)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = prepareInitRequest(requestBody, w)
	//if err != nil {
	//	return err
	//}
	//
	//err = checkPreviousInitialization(w)
	//if err != nil {
	//	return err
	//}
	//
	//// The existence of an admin signing token also means the system is
	//// initialized.
	//log.Log().Info("routeInit", "msg", "No admin signing token. will create one")
	//
	//adminSigningTokenBytes, err := generateAdminSigningToken(w)
	//if err != nil {
	//	return err
	//}
	//
	//// Generate salt to hash the recovery token.
	////
	//// The recovery token is a secure passphrase that we send to the admin user
	//// upon first time initialization. They can use this token to recover
	//// the root key in case the system crashes and there is an encrypted backup
	//// of the root key.
	////
	//// Since it acts like a password, we salt and store its hash accordingly.
	//// When the admin provides the recovery token, we verify it by hashing it
	//// with the same salt and comparing it to the stored hash. If they match,
	//// we can decrypt the root key and re-key the system.
	//salt, err := generateSalt(w)
	//if err != nil {
	//	return err
	//}
	//
	//recoveryToken := crypto.Token()
	//
	//err = updateStateForInit(recoveryToken, adminSigningTokenBytes, salt)
	//if err != nil {
	//	return err
	//}
	//
	//responseBody := net.MarshalBody(reqres.InitResponse{
	//	RecoveryToken: recoveryToken,
	//}, w)
	//if responseBody == nil {
	//	return errors.New("failed to marshal response body")
	//}
	//
	//net.Respond(http.StatusOK, responseBody, w)
	//log.Log().Info("routeInit", "msg", "OK")
	//return nil
}
