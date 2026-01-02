//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// guardSecretRequest is a generic helper that validates secret requests by
// performing authentication, authorization, and path validation. It extracts
// the common validation pattern used across secret operations (get, put,
// delete, undelete, etc.).
//
// On failure, this function automatically writes the appropriate HTTP error
// response before returning the error.
//
// Type Parameters:
//   - TUnauth: The response type for unauthorized access errors
//   - TBadInput: The response type for invalid path errors
//
// Parameters:
//   - path: The namespace path to validate and authorize
//   - permissions: The required permissions for the operation
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//   - unauthorizedResp: The error response to send if unauthorized
//   - badInputResp: The error response to send if the path is invalid
//
// Returns:
//   - *sdkErrors.SDKError: An error if authentication, authorization, or
//     validation fails. Returns nil if all validations pass.
func guardSecretRequest[TUnauth, TBadInput any](
	path string,
	permissions []data.PolicyPermission,
	w http.ResponseWriter,
	r *http.Request,
	unauthorizedResp TUnauth,
	badInputResp TBadInput,
) *sdkErrors.SDKError {
	// Extract and validate peer SPIFFE ID
	peerSPIFFEID, err := net.ExtractPeerSPIFFEIDFromRequestAndRespondOnFail[TUnauth](
		r, w, unauthorizedResp,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	// Check access permissions
	allowed := state.CheckAccess(peerSPIFFEID.String(), path, permissions)
	if !allowed {
		failErr := net.Fail(unauthorizedResp, w, http.StatusUnauthorized)
		if failErr != nil {
			return sdkErrors.ErrAccessUnauthorized.Wrap(failErr)
		}
		return sdkErrors.ErrAccessUnauthorized
	}

	// Validate path format
	pathErr := validation.ValidatePath(path)
	if pathErr != nil {
		failErr := net.Fail(badInputResp, w, http.StatusBadRequest)
		pathErr.Msg = "invalid secret path: " + path
		if failErr != nil {
			return pathErr.Wrap(failErr)
		}
		return pathErr
	}

	return nil
}
