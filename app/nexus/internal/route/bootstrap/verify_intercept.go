//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/auth"

	"github.com/spiffe/spike/internal/net"
)

func guardVerifyRequest(
	request reqres.BootstrapVerifyRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.BootstrapVerifyResponse](
		r, w, reqres.BootstrapVerifyResponse{
			Err: data.ErrUnauthorized,
		})
	if err != nil {
		return err
	}

	if !spiffeid.IsBootstrap(peerSPIFFEID.String()) {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrUnauthorized,
			}, w)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	// Validate nonce size (AES-GCM standard nonce is 12 bytes)
	const expectedNonceSize = 12 // TODO: to constants.
	if len(request.Nonce) != expectedNonceSize {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrBadInput,
			}, w)
		if err == nil {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	// Validate ciphertext size to prevent DoS attacks
	const maxCiphertextSize = 1024 // TODO: to constants.
	if len(request.Ciphertext) > maxCiphertextSize {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.BootstrapVerifyResponse{
				Err: data.ErrBadInput,
			}, w)
		if err == nil {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	return nil
}
