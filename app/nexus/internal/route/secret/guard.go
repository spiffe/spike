//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	stdErrs "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// guardSecretRequest is a generic helper that validates secret requests by
// performing authentication, authorization, and path validation.
//
// This function extracts the common validation pattern used across secret
// operations (get, put, delete, undelete, etc.).
//
// Parameters:
//   - path: The secret path to validate and authorize
//   - permissions: The required permissions for the operation
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//   - unauthorizedResp: The error response to send if unauthorized
//   - badInputResp: The error response to send if the path is invalid
//   - fName: The function name for logging
//
// Returns:
//   - nil if all validations pass
//   - error if authentication, authorization, or validation fails
func guardSecretRequest[TUnauth, TBadInput any](
	path string,
	permissions []data.PolicyPermission,
	w http.ResponseWriter,
	r *http.Request,
	unauthorizedResp TUnauth,
	badInputResp TBadInput,
	fName string,
) error {
	// Extract and validate peer SPIFFE ID
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[TUnauth](
		r, w, unauthorizedResp,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Validate path format
	err = validation.ValidatePath(path)
	if err != nil {
		failErr := stdErrs.Join(sdkErrors.ErrInvalidInput, err)
		net.Fail(badInputResp, w, http.StatusBadRequest)
		return failErr
	}

	// Check access permissions
	allowed := state.CheckAccess(
		peerSPIFFEID.String(),
		path,
		permissions,
	)
	if !allowed {
		net.Fail(unauthorizedResp, w, http.StatusUnauthorized)
		return sdkErrors.ErrUnauthorized
	}

	return nil
}
