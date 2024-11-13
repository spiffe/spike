//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spiffe/spike/internal/log"
)

// CustomClaims embeds RegisteredClaims to inherit standard JWT fields
type CustomClaims struct {
	AdminTokenID string `json:"adminTokenId"`
	// Embed the standard claims
	*jwt.RegisteredClaims
}

// ValidateJwt checks if a request has a valid JSON Web Token (JWT).
//
// This function validates the JWT token from the request against the provided
// admin token. If validation fails, it sends a 401 Unauthorized response with
// a JSON error message.
//
// Parameters:
//   - w: http.ResponseWriter - The response writer for error handling
//   - r: *http.Request - The incoming request containing the JWT
//   - adminToken: string - The admin token to validate against
//
// Returns:
//   - bool - true if the JWT is valid, false otherwise
//
// Response on failure:
//   - Status: 401 Unauthorized
//   - Content-Type: application/json
//   - Body: {"status": "unauthorized"}
func ValidateJwt(w http.ResponseWriter, r *http.Request, adminToken string) bool {
	// Extract JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		log.Log().Info(
			"ValidateJwt",
			"msg", "Missing or invalid authorization header",
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, `{"status": "unauthorized"}`)
		if err != nil {
			log.Log().Error(
				"ValidateJwt",
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

		log.Log().Info("ValidateJwt",
			"msg", "Invalid token",
			"err", errorString)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err := io.WriteString(w, `{"status": "unauthorized"}`)
		if err != nil {
			log.Log().Error("ValidateJwt",
				"msg", "Problem writing response",
				"err", err.Error())
		}

		return false
	}

	// Validate custom claims
	// TODO: there are magic strings here. Replace with consts.
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

	log.Log().Info("ValidateJwt", "msg", "Valid token")

	return true
}

// CreateJwt generates a new JWT token for admin authentication.
//
// The function creates a JWT with the following characteristics:
//   - Expires in 24 hours from creation
//   - Issued by "spike-nexus"
//   - Subject is "spike-admin"
//   - Uses HS256 signing method
//   - Includes custom AdminTokenID claim
//
// Parameters:
//   - adminToken: string - The token used to sign the JWT
//   - w: http.ResponseWriter - The response writer for error handling
//
// Returns:
//   - string - The signed JWT token, or empty string if signing fails
//
// On signing failure:
//   - Writes 500 Internal Server Error status
//   - Logs the error
//   - Returns empty string
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
		_, err = io.WriteString(w, `{"error":"internal server error"}`)
		if err != nil {
			log.Log().Error("routeAdminLogin",
				"msg", "Problem writing response", "err", err.Error())
		}
		return ""
	}

	return signedToken
}
