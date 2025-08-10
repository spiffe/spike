//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardPutSecretMetadataRequest(
	request reqres.SecretPutRequest, w http.ResponseWriter, r *http.Request,
) error {
	values := request.Values
	path := request.Path

	spiffeid, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if spiffeid == nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidatePath(path)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	for k := range values {
		err := validation.ValidateName(k)
		if err != nil {
			responseBody := net.MarshalBody(reqres.SecretPutResponse{
				Err: data.ErrBadInput,
			}, w)
			net.Respond(http.StatusUnauthorized, responseBody, w)
			return apiErr.ErrInvalidInput
		}
	}

	allowed := state.CheckAccess(
		spiffeid.String(), path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
