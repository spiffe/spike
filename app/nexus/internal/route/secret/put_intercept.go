//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/internal/auth"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardSecretPutRequest(
	request reqres.SecretPutRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.SecretPutResponse](
		r, w, reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	values := request.Values
	path := request.Path

	err = validation.ValidatePath(path)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretPutResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	for k := range values {
		err := validation.ValidateName(k)
		if err != nil {
			responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
				reqres.SecretPutResponse{
					Err: data.ErrBadInput,
				}, w)
			alreadyResponded = err != nil
			if !alreadyResponded {
				net.Respond(http.StatusUnauthorized, responseBody, w)
			}
			return apiErr.ErrInvalidInput
		}
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretPutResponse{
				Err: data.ErrUnauthorized,
			}, w)
		respondedAlready := err != nil // TODO: add this pattern to all usages.
		if !respondedAlready {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	return nil
}
