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
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

func guardSecretUndeleteRequest(
	request reqres.SecretUndeleteRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.SecretUndeleteResponse](
		r, w, reqres.SecretUndeleteResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil // TODO: copy this pattern to all guards.
	if alreadyResponded {
		return err
	}

	path := request.Path
	err = validation.ValidatePath(path)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretUndeleteResponse{
				Err: data.ErrBadInput,
			}, w)
		if err == nil {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return errors.ErrInvalidInput
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(),
		path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretUndeleteResponse{
				Err: data.ErrUnauthorized,
			}, w)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return errors.ErrUnauthorized
	}

	return nil
}
