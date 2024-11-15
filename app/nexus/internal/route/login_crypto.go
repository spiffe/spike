//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

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

func decodePasswordHash(passwordHash string, w http.ResponseWriter) ([]byte, error) {
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
