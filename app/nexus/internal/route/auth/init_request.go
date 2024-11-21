//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
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
