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
	"io"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeInit(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInit",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.InitRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Info("routeInit",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	password := req.Password

	if len(password) < 16 {
		w.WriteHeader(http.StatusBadRequest)

		res := reqres.InitResponse{
			Err: reqres.ErrLowEntropy,
		}
		body, err := json.Marshal(res)
		if err != nil {
			res.Err = reqres.ErrServerFault

			log.Log().Error("routeInit",
				"msg", "Problem generating response", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	adminToken := state.AdminToken()
	if adminToken != "" {
		log.Log().Info("routeInit",
			"msg", "Already initialized")

		w.WriteHeader(http.StatusBadRequest)

		res := reqres.InitResponse{
			Err: reqres.ErrAlreadyInitialized,
		}
		body, err := json.Marshal(res)
		if err != nil {
			res.Err = reqres.ErrServerFault

			log.Log().Error("routeInit",
				"msg", "Problem generating response", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem writing response", "err", err.Error())
		}

		return
	}

	// No admin token yet.

	// Generate adminToken (32 bytes)
	adminTokenBytes := make([]byte, 32)
	if _, err := rand.Read(adminTokenBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Log().Error("routeInit",
			"msg", "Failed to generate admin token", "err", err.Error())

		res := reqres.InitResponse{
			Err: reqres.ErrServerFault,
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem generating response", "err", err.Error())
		}

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem writing response", "err", err.Error())
		}

		return
	}

	// Generate salt and hash password
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Log().Error("routeInit",
			"msg", "Failed to generate salt", "err", err.Error())

		res := reqres.InitResponse{
			Err: reqres.ErrServerFault,
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem generating response", "err", err.Error())
		}

		_, err = io.WriteString(w, string(body))
		if err != nil {
			log.Log().Error("routeInit",
				"msg", "Problem writing response", "err", err.Error())
		}

		return
	}

	iterationCount := 600_000 // Minimum OWASP recommendation for PBKDF2-SHA256
	hashLength := 32          // 256 bits output

	passwordHash := pbkdf2.Key([]byte(password), salt,
		iterationCount, hashLength, sha256.New)

	state.SetAdminToken("spike." + string(adminTokenBytes))
	state.SetAdminCredentials(
		hex.EncodeToString(passwordHash),
		hex.EncodeToString(salt),
	)

	w.WriteHeader(http.StatusOK)

	res := reqres.InitResponse{}
	body, err := json.Marshal(res)
	if err != nil {
		res.Err = reqres.ErrServerFault

		log.Log().Error("routeInit",
			"msg", "Problem generating response", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = io.WriteString(w, string(body))
	if err != nil {
		log.Log().Error("routeInit",
			"msg", "Problem writing response", "err", err.Error())
	}
}
