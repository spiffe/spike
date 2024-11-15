//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"errors"
	"github.com/spiffe/spike/app/nexus/internal/config"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

func prepareInitRequestBody(
	w http.ResponseWriter, r *http.Request,
) ([]byte, error) {
	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return []byte{}, errors.New("failed to read request body")
	}
	return requestBody, nil
}

func prepareInitRequest(
	requestBody []byte, w http.ResponseWriter,
) (*reqres.InitRequest, error) {
	request := net.HandleRequest[
		reqres.InitRequest, reqres.InitResponse](
		requestBody, w,
		reqres.InitResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return nil, errors.New("failed to parse request body")
	}
	return request, nil
}

func sanitizeInitRequest(
	req *reqres.InitRequest, w http.ResponseWriter,
) (*reqres.InitRequest, error) {
	password := req.Password
	if len(password) < config.SpikeNexusAdminPasswordMinLength {
		res := reqres.InitResponse{Err: reqres.ErrLowEntropy}

		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return nil, errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		log.Log().Info("routeInit", "msg", "exit: Password too short")
		return nil, errors.New("password too short")
	}

	return req, nil
}
