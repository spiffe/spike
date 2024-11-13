//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spiffe/spike/internal/log"
	"io"
	"net/http"
	"strings"
	"time"
)

// CustomClaims embeds RegisteredClaims to inherit standard JWT fields
type CustomClaims struct {
	AdminTokenID string `json:"adminTokenId"`
	// Embed the standard claims
	*jwt.RegisteredClaims
}

func ensureValidJwt(w http.ResponseWriter, r *http.Request, adminToken string) bool {
	// Extract JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		log.Log().Info(
			"ensureValidJwt",
			"msg", "Missing or invalid authorization header",
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, `{"status": "unauthorized"}`)
		if err != nil {
			log.Log().Error(
				"ensureValidJwt",
				"msg", "Problem writing response",
				"err", err.Error(),
			)
		}

		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(
		tokenString, &CustomClaims{},
		func(token *jwt.Token) (any, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil,
					fmt.Errorf(
						"unexpected signing method: %v", token.Header["alg"])
			}

			if adminToken == "" {
				return nil, fmt.Errorf("admin token not found")
			}

			// Return the key used to sign the token
			return []byte(adminToken), nil
		})

	if err != nil || !token.Valid {
		errorString := ""
		if err != nil {
			errorString = err.Error()
		}

		log.Log().Info("ensureValidJwt",
			"msg", "Invalid token",
			"err", errorString)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, `{"status": "unauthorized"}`)
		if err != nil {
			log.Log().Error("ensureValidJwt",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return false
	}

	// Validate custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || claims.Issuer != "spike-nexus" ||
		claims.AdminTokenID != "spike-admin-jwt" {
		log.Log().Info("routeGetSecret",
			"msg", "Invalid token claims")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, `{"status": "unauthorized"}`)
		if err != nil {
			log.Log().Error("routeGetSecret",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return false
	}

	log.Log().Info("routeGetSecret", "msg", "Valid token")

	return true
}

func ValidateJwt(w http.ResponseWriter, r *http.Request, adminToken string) bool {
	if ensureValidJwt(w, r, adminToken) {
		return true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_, err := io.WriteString(w, `{"status": "unauthorized"}`)
	if err != nil {
		errorString := err.Error()

		log.Log().Error("routeDeleteSecret",
			"msg", "Problem writing response",
			"err", errorString)
	}

	return false
}

func CreateJwt(adminToken string, w http.ResponseWriter) string {
	// Create JWT with claims
	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "spike-nexus", // TODO: to consts.
			Subject:   "spike-admin", // TODO: to consts
		},
		AdminTokenID: "spike-admin-jwt", // TODO: to consts
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
		return ""
	}

	return signedToken
}
