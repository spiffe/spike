//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"strings"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func ensureValidJwt(w http.ResponseWriter, r *http.Request) bool {
	// Extract JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		log.Log().Info("routeGetSecret", "msg", "Missing or invalid authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response", "err", err.Error())
		}
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &state.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		adminToken := state.AdminToken()

		// Return the key used to sign the token
		return []byte(adminToken), nil
	})

	if err != nil || !token.Valid {
		log.Log().Info("routeGetSecret",
			"msg", "Invalid token",
			"err", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response", "err", err.Error())
		}
		return false
	}

	// Validate custom claims
	claims, ok := token.Claims.(*state.CustomClaims)
	if !ok || claims.Issuer != "nexus" || claims.AdminTokenID != "spike-admin-jwt" {
		log.Log().Info("routeGetSecret", "msg", "Invalid token claims")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response", "err", err.Error())
		}
		return false
	}

	log.Log().Info("routeGetSecret", "msg", "Valid token")

	return true
}

func routeGetSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeGetSecret",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := ensureValidJwt(w, r)
	if !validJwt {
		return
	}

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.SecretReadRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem unmarshalling request",
			"err", err.Error())
		return
	}

	version := req.Version
	path := req.Path

	secret, exists := state.GetSecret(path, version)
	if !exists {
		log.Log().Info("routeGetSecret", "msg", "Secret not found")
		w.WriteHeader(http.StatusNotFound)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response", "err", err.Error())
		}
		return
	}

	res := reqres.SecretReadResponse{Data: secret}
	md, err := json.Marshal(res)
	if err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem generating response", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Log().Info("routeGetSecret", "msg", "Got secret")

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Log().Error("routeGetSecret",
			"msg", "Problem writing response", "err", err.Error())
	}
}
