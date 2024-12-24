//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardGetSecretMetadataRequest(
	request reqres.SecretMetadataRequest, w http.ResponseWriter, r *http.Request,
) error {
	path := request.Path

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretMetadataResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretMetadataResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
	}

	err = validation.ValidatePath(path)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretMetadataResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	allowed := state.CheckAccess(
		spiffeid.String(), path,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretListResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return sdkErr.ErrUnauthorized
	}

	return nil
}
