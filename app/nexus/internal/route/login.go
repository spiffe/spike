//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"golang.org/x/crypto/pbkdf2"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeAdminLogin(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeAdminLogin", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	// TODO: signature should be `w, r` for consistency.
	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.AdminLoginRequest, reqres.AdminLoginResponse](
		requestBody, w,
		reqres.AdminLoginResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	password := request.Password
	creds := state.AdminCredentials()
	passwordHash := creds.PasswordHash
	salt := creds.Salt

	s, err := hex.DecodeString(salt)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding salt",
			"err", err.Error())

		body := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault,
		}, w)
		if body == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, body, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return
	}

	// TODO: duplication.
	// TODO: make this configurable.
	iterationCount := 600_000 // Minimum OWASP recommendation for PBKDF2-SHA256
	hashLength := 32          // 256 bits output

	ph := pbkdf2.Key(
		[]byte(password), s,
		iterationCount, hashLength, sha256.New,
	)

	b, err := hex.DecodeString(passwordHash)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding password hash",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "OK")
		return
	}

	if !hmac.Equal(ph, b) {
		log.Log().Info("routeAdminLogin", "msg", "Invalid password")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrUnauthorized}, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusUnauthorized, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return
	}

	adminToken := state.AdminToken()
	if adminToken == "" {
		log.Log().Error("routeAdminLogin", "msg", "Admin token not set")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return
	}

	signedToken := net.CreateJwt(adminToken, w)
	if signedToken == "" {
		return
	}

	responseBody := net.MarshalBody(reqres.AdminLoginResponse{
		Token: signedToken,
	}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeAdminLogin", "msg", "authorized")
}
