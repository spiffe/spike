//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"net/http"
	"time"
)

// TODO: may be used elsewhere.
func signToken(token, adminToken []byte, metadata state.TokenMetadata) []byte {
	h := hmac.New(sha256.New, adminToken)
	h.Write(token)
	metadataBytes, _ := json.Marshal(metadata)
	h.Write(metadataBytes)
	return h.Sum(nil)
}

func routeAdminLogin(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeAdminLogin",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	// TODO: signature should be `w, r` for consistency.
	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var request reqres.AdminLoginRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &request)); err != nil {
		log.Log().Info("routeAdminLogin",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	password := request.Password

	creds := state.AdminCredentials()
	salt := creds.Salt
	passwordHash := creds.PasswordHash

	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> GOT ADMIN CREDENTIALS")
	fmt.Println("SALT: ", salt)
	fmt.Println("PASSWORD HASH: ", passwordHash)

	s, err := hex.DecodeString(salt)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding salt",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	// TODO: duplication.
	// TODO: make this configurable.
	iterationCount := 600_000 // Minimum OWASP recommendation for PBKDF2-SHA256
	hashLength := 32          // 256 bits output

	ph := pbkdf2.Key([]byte(password), s, iterationCount, hashLength, sha256.New)

	b, err := hex.DecodeString(passwordHash)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding password hash",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	if !hmac.Equal(ph, b) {
		log.Log().Info("routeAdminLogin",
			"msg", "Invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	// TODO: adminToken / not initialized checks come to these
	// functions now. it's better to send an error message/code instead.
	// SPIKE Pilot can parse that code and return a proper error message.

	adminToken := state.AdminToken()
	if adminToken == "" {
		log.Log().Error("routeAdminLogin",
			"msg", "Admin token not set")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	const tokenLength = 32
	// Generate session token
	token := make([]byte, tokenLength)
	if _, err := rand.Read(token); err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Failed to generate session token",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	// Create JWT with claims
	now := time.Now()
	claims := state.CustomClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nexus",
			Subject:   "admin", // or user ID
		},
		AdminTokenID: "spike-admin-jwt",
	}

	// Create token with claims
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with admin token
	signedToken, err := tok.SignedString([]byte(adminToken))
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Failed to sign token",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	res := reqres.AdminLoginResponse{
		Token: signedToken,
	}
	body, err = json.Marshal(res)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Failed to marshal response",
			"err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem writing response", "err", err.Error())
	}
}
