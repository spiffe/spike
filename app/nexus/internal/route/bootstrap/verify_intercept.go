//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/internal/net"
)

func guardVerifyRequest(
	request reqres.BootstrapVerifyRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.BootstrapVerifyResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if peerSPIFFEID == nil {
		responseBody := net.MarshalBody(reqres.BootstrapVerifyResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if !spiffeid.IsBootstrap(peerSPIFFEID.String()) {
		responseBody := net.MarshalBody(reqres.BootstrapVerifyResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	// Validate nonce size (AES-GCM standard nonce is 12 bytes)
	const expectedNonceSize = 12
	if len(request.Nonce) != expectedNonceSize {
		responseBody := net.MarshalBody(reqres.BootstrapVerifyResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	// Validate ciphertext size to prevent DoS attacks
	const maxCiphertextSize = 1024
	if len(request.Ciphertext) > maxCiphertextSize {
		responseBody := net.MarshalBody(reqres.BootstrapVerifyResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	return nil
}
