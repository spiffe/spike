//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spiffe/spike/internal/entity/data"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func routeInitCheck(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeInitCheck",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.CheckInitStateRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Info("routeInitCheck",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	adminToken := state.AdminToken()

	if adminToken != "" {
		log.Log().Info("routeInitCheck",
			"msg", "Already initialized")

		res := reqres.CheckInitStateResponse{
			State: data.AlreadyInitialized,
		}
		md, err := json.Marshal(res)
		if err != nil {
			log.Log().Error("routeInitCheck",
				"msg", "Problem generating response", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		_, err = io.WriteString(w, string(md))
		if err != nil {
			log.Log().Error("routeInitCheck",
				"msg", "Problem writing response", "err", err.Error())
		}

		return
	}

	res := reqres.CheckInitStateResponse{
		State: data.NotInitialized,
	}
	md, err := json.Marshal(res)

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeInitCheck",
			"msg", "Problem writing response", "err", err.Error())
	}

}

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
		_, err := io.WriteString(w, "")
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
		_, err := io.WriteString(w, "")
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

		_, err := io.WriteString(w, "")
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

		_, err := io.WriteString(w, "")
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

	// Good first issue.
	// TODO: spike init asks to save password on the machine.
	// will save to ~/.spike/credentials.json
	// configurable.
	// if the file does not exist, it will ask for password.
	// will also remind the risks of doing so.
	// or maybe not worth it because people are lazy.

	fmt.Println(">>>>>>>>>>>>>>>> SETTING ADMIN CREDENTIALS")
	fmt.Println("passwordHash: ", string(passwordHash))
	fmt.Println("salt: ", string(salt))

	state.SetAdminCredentials(hex.EncodeToString(passwordHash), hex.EncodeToString(salt))

	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "")
	if err != nil {
		log.Log().Error("routeInit",
			"msg", "Problem writing response", "err", err.Error())
	}
}
