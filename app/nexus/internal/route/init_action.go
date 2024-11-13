//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"golang.org/x/crypto/pbkdf2"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func newInitRequest(
	requestBody []byte, w http.ResponseWriter,
) *reqres.InitRequest {
	var request reqres.InitRequest
	if err := net.HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("newInitRequest",
			"msg", "Problem unmarshalling request",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.InitResponse{
			Err: reqres.ErrBadInput}, w)
		if responseBody == nil {
			return nil
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}
	return &request
}

func routeInit(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInit",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := newInitRequest(requestBody, w)
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
		return
	}

	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInit",
			"msg", "Already initialized")

		res := reqres.InitResponse{
			Err: reqres.ErrAlreadyInitialized,
		}

		responseBody := net.MarshalBody(res, w)
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

		res := reqres.InitResponse{
			Err: reqres.ErrServerFault,
		}

		responseBody := net.MarshalBody(res, w)
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
		log.Log().Error("routeInit",
			"msg", "Failed to generate salt",
			"err", err.Error())

		res := reqres.InitResponse{
			Err: reqres.ErrServerFault,
		}

		responseBody := net.MarshalBody(res, w)
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

	passwordHash := pbkdf2.Key([]byte(password), salt,
		iterationCount, hashLength, sha256.New)

	state.SetAdminToken("spike." + string(adminTokenBytes))
	state.SetAdminCredentials(
		hex.EncodeToString(passwordHash),
		hex.EncodeToString(salt),
	)

	res := reqres.InitResponse{}

	responseBody := net.MarshalBody(res, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeInit", "msg", "OK")
}
