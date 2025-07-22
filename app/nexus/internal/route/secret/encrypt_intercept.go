//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardEncryptSecretRequest(
	_ reqres.SecretEncryptRequest, w http.ResponseWriter, r *http.Request,
) error {
	// TODO: some of these flows can be factored out if we keep the `request`
	// a generic parameter. That will deduplicate some of the validation code.

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	allowed := state.CheckAccess(
		spiffeid.String(),
		"spike/system/secrets/encrypt",
		[]data.PolicyPermission{data.PermissionExecute},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
