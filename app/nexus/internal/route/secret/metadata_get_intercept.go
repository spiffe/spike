//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// guardGetSecretMetadataRequest validates a secret metadata retrieval request
// by performing authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the secret path format
//   - Checks if the peer has read permission for the specified secret path
//
// Read permission is required to retrieve secret metadata. The authorization
// check is performed against the specific secret path to enable fine-grained
// access control. Metadata access uses the same permission level as secret
// data access.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The secret metadata request containing the secret path
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: An error if authentication, authorization, or path
//     validation fails. Returns nil if all validations pass.
func guardGetSecretMetadataRequest(
	request reqres.SecretMetadataRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := net.ExtractPeerSPIFFEIDFromRequestAndRespondOnFail[reqres.SecretMetadataResponse](
		r, w, reqres.SecretMetadataResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path
	pathErr := validation.ValidatePath(path)
	if pathErr != nil {
		failErr := net.Fail(
			reqres.SecretMetadataResponse{}.BadRequest(), w,
			http.StatusBadRequest,
		)
		pathErr.Msg = "invalid secret path: " + path
		if failErr != nil {
			return pathErr.Wrap(failErr)
		}
		return pathErr
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), path,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		failErr := net.Fail(
			reqres.SecretMetadataResponse{}.Unauthorized(), w,
			http.StatusUnauthorized,
		)
		authErr := sdkErrors.ErrAccessUnauthorized.Clone()
		authErr.Msg = "unauthorized to read secret metadata for: " + path
		if failErr != nil {
			return authErr.Wrap(failErr)
		}
		return authErr
	}

	return nil
}
