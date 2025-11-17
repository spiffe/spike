//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
)

// guardDeleteSecretRequest validates a secret deletion request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the secret path format
//   - Checks if the peer has write permission for the specified secret path
//
// Write permission is required for delete operations following the principle
// that deletion is a write operation on the secret resource.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The secret deletion request containing the secret path
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if path validation fails
func guardDeleteSecretRequest(
	request reqres.SecretDeleteRequest, w http.ResponseWriter, r *http.Request,
) error {
	return guardSecretRequest(
		request.Path,
		[]data.PolicyPermission{data.PermissionWrite},
		w, r,
		reqres.SecretDeleteUnauthorized,
		reqres.SecretDeleteBadInput,
		"guardDeleteSecretRequest",
	)
}
