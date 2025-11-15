//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/internal/auth"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardPolicyReadRequest(
	request reqres.PolicyReadRequest, w http.ResponseWriter, r *http.Request,
) error {
	resUnauthorized := reqres.PolicyReadResponse{Err: data.ErrUnauthorized}
	resBadInput := reqres.PolicyReadResponse{Err: data.ErrBadInput}

	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.PolicyReadResponse](
		r, w, resUnauthorized)
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	policyID := request.ID

	err = validation.ValidatePolicyID(policyID)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			resBadInput, w,
		)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), apiAuth.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			resUnauthorized, w,
		)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	return nil
}
