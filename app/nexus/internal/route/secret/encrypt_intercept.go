//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardEncryptSecretRequest(
	_ reqres.SecretEncryptRequest, w http.ResponseWriter, r *http.Request,
) error {
	// TODO: some of these flows can be factored out if we keep the `request`
	// a generic parameter. That will deduplicate some of the validation code.

	sid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSpiffeId(sid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	// TODO: this check should have been within state.CheckAccess
	// maybe we can fork based on spike/system/secrets/encrypt.
	//
	// Lite Workloads are always allowed:
	allowed := false
	if spiffeid.IsLiteWorkload(
		env.TrustRootForLiteWorkload(), sid.String()) {
		allowed = true
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		allowed = state.CheckAccess(
			sid.String(),
			"spike/system/secrets/encrypt",
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}
	// If not, block the request:
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretEncryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
