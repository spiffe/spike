//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/errors"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/internal/net"
)

func guardSecretUndeleteRequest(
	request reqres.SecretUndeleteRequest, w http.ResponseWriter, r *http.Request,
) error {
	path := request.Path

	sid, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	if sid == nil {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(sid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	err = validation.ValidatePath(path)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return errors.ErrInvalidInput
	}

	allowed := state.CheckAccess(
		sid.String(),
		path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretUndeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	return nil
}
