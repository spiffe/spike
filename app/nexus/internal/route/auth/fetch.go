package auth

import (
	"errors"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// fetchAdminToken retrieves the administrator authentication token from the
// application state.
// If the token is not set, it handles the error by responding with an
// appropriate HTTP status code and error message.
//
// Parameters:
//   - w: HTTP ResponseWriter for sending error responses
//
// Returns:
//   - string: The administrator token if successfully retrieved
//   - error: nil if the token exists, or an error if the token is not set or
//     response marshaling fails
//
// The function will respond with HTTP 500 Internal Server Error and return an
// empty string if the admin token is not set in the application state.
//
// Note: The function logs server errors using the application's logging system.
func fetchAdminToken(w http.ResponseWriter) (string, error) {
	adminToken := state.AdminSigningToken()
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
