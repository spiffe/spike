//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

func guardDeletePolicyRequest(
	request reqres.PolicyDeleteRequest, w http.ResponseWriter, r *http.Request,
) error {
	policyId := request.Id

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
	}

	err = validation.ValidatePolicyId(policyId)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}

		net.Respond(http.StatusBadRequest, responseBody, w)
		return err
	}

	allowed := state.CheckAccess(
		spiffeid.String(), "*",
		[]data.PolicyPermission{data.PermissionSuper},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return errors.ErrUnauthorized
	}

	return nil
}
