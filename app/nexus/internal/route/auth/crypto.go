package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// checkHmac verifies if the provided password hash matches the expected hash
// using HMAC comparison. It handles unauthorized access attempts by responding
// with appropriate HTTP status codes and error messages.
//
// Parameters:
//   - ph: The provided password hash as a byte slice
//   - b: The expected password hash to compare against
//   - w: HTTP ResponseWriter for sending error responses
//
// Returns:
//   - error: nil if the HMAC verification succeeds, or an error describing
//     the failure
//
// The function will return an "invalid password" error and respond with
// HTTP 401 Unauthorized if the HMAC comparison fails.
func checkHmac(ph, b []byte, w http.ResponseWriter) error {
	if !hmac.Equal(ph, b) {
		log.Log().Info("routeAdminLogin", "msg", "Invalid password")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrUnauthorized}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusUnauthorized, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return errors.New("invalid password")
	}
	return nil
}

// decodePasswordHash decodes a hexadecimal password hash string into bytes.
// It handles decoding errors by responding with appropriate HTTP status
// codes and error messages.
//
// Parameters:
//   - passwordHash: Hexadecimal string representation of the password hash
//   - w: HTTP ResponseWriter for sending error responses
//
// Returns:
//   - []byte: Decoded password hash bytes
//   - error: nil if decoding succeeds, or an error describing the failure
//
// The function will respond with HTTP 500 Internal Server Error and return
// an empty byte slice if the decoding fails or if response marshaling fails.
func decodePasswordHash(
	passwordHash string, w http.ResponseWriter,
) ([]byte, error) {
	b, err := hex.DecodeString(passwordHash)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding password hash",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return []byte{}, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "OK")
		return []byte{}, errors.New("failed to decode password hash")
	}
	return b, nil
}

// decodeSalt decodes a hexadecimal salt string into bytes.
// It handles decoding errors by responding with appropriate HTTP status codes
// and error messages.
//
// Parameters:
//   - salt: Hexadecimal string representation of the salt
//   - w: HTTP ResponseWriter for sending error responses
//
// Returns:
//   - []byte: Decoded salt bytes
//   - error: nil if decoding succeeds, or an error describing the failure
//
// The function will respond with HTTP 500 Internal Server Error and return an
// empty byte slice if the decoding fails or if response marshaling fails.
//
// Note: The function logs unauthorized access attempts and server errors
// using the application's logging system.
func decodeSalt(salt string, w http.ResponseWriter) ([]byte, error) {
	s, err := hex.DecodeString(salt)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding salt",
			"err", err.Error())

		body := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault,
		}, w)
		if body == nil {
			return []byte{}, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, body, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return []byte{}, errors.New("failed to decode salt")
	}
	return s, nil
}

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
			Err: reqres.ErrServerFault}, w,
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
