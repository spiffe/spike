//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	cfg "github.com/spiffe/spike-sdk-go/config/auth"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// guardPolicyCreateRequest validates a policy creation request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer has write permission for the policy access path
//   - Validates the policy name format
//   - Validates the SPIFFE ID pattern (regex)
//   - Validates the path pattern (regex)
//   - Validates the permissions list
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy creation request containing policy details
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if any input validation fails
func guardPolicyCreateRequest(
	request reqres.PolicyCreateRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.PolicyCreateResponse](
		r, w, reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	// Request "write" access to the ACL system for the SPIFFE ID.
	allowed := state.CheckAccess(
		peerSPIFFEID.String(), cfg.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.PolicyCreateResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	name := request.Name
	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	err = validation.ValidateName(name)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidateSPIFFEIDPattern(SPIFFEIDPattern)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidatePathPattern(pathPattern)
	if err != nil {
		responseBody, err :=
			net.MarshalBodyAndRespondOnMarshalFail(reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidatePermissions(permissions)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	return nil
}
