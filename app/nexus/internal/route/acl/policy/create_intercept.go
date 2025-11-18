//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	cfg "github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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
	const fName = "guardPolicyCreateRequest"

	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.PolicyCreateResponse](
		r, w, reqres.PolicyCreateUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Request "write" access to the ACL system for the SPIFFE ID.
	allowed := state.CheckAccess(
		peerSPIFFEID.String(), cfg.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		return net.Fail(
			reqres.PolicyCreateUnauthorized, w,
			http.StatusUnauthorized, sdkErrors.ErrUnauthorized, fName,
		)
	}

	name := request.Name
	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	if err := validation.ValidateName(name); err != nil {
		return net.Fail(
			reqres.PolicyCreateBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	if err := validation.ValidateSPIFFEIDPattern(SPIFFEIDPattern); err != nil {
		return net.Fail(
			reqres.PolicyCreateBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	if err := validation.ValidatePathPattern(pathPattern); err != nil {
		return net.Fail(
			reqres.PolicyCreateBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	if err := validation.ValidatePermissions(permissions); err != nil {
		return net.Fail(
			reqres.PolicyCreateBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	return nil
}
