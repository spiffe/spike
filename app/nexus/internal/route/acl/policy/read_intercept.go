//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/config/auth"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardReadPolicyRequest(
	request reqres.PolicyReadRequest, w http.ResponseWriter, r *http.Request,
) error {

	peerSPIFFEID, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	if peerSPIFFEID == nil {
		responseBody := net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(peerSPIFFEID.String())
	if err != nil {
		responseBody := net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	policyID := request.ID

	err = validation.ValidatePolicyID(policyID)
	if err != nil {
		responseBody := net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyReadResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		return apiErr.ErrInvalidInput
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), auth.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		responseBody := net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyReadResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	return nil
}
