//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardDecryptSecretRequest(
	_ reqres.SecretDecryptRequest, w http.ResponseWriter, r *http.Request,
) error {
	sid, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if sid == nil {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(sid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
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
			"spike/system/secrets/decrypt",
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
