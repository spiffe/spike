//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// guardSecretUndeleteRequest validates a secret restoration request by
// performing authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer has write permission for the specified secret path
//   - Validates the secret path format
//
// Write permission is required for undelete operations following the principle
// that restoration is a write operation on the secret resource. The
// authorization check is performed against the specific secret path to enable
// fine-grained access control.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter, and an error is returned.
//
// Parameters:
//   - request: The secret undelete request containing the secret path
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - sdkErrors.ErrAccessUnauthorized if authorization fails
//   - sdkErrors.ErrAPIBadRequest if path validation fails
//   - SDK errors from authentication if peer SPIFFE ID extraction fails
func guardSecretUndeleteRequest(
	request reqres.SecretUndeleteRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	return guardSecretRequest(
		request.Path,
		[]data.PolicyPermission{data.PermissionWrite},
		w, r,
		reqres.SecretUndeleteResponse{}.Unauthorized(),
		reqres.SecretUndeleteResponse{}.BadRequest(),
	)
}
