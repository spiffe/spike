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

// guardGetSecretRequest validates a secret retrieval request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the secret path format
//   - Checks if the peer has read permission for the specified secret path
//
// Read permission is required to retrieve secret data. The authorization check
// is performed against the specific secret path to enable fine-grained access
// control.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The secret read request containing the secret path
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - *sdkErrors.SDKError: An error if authentication, authorization, or path
//     validation fails. Returns nil if all validations pass.
func guardGetSecretRequest(
	request reqres.SecretGetRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	return guardSecretRequest(
		request.Path,
		[]data.PolicyPermission{data.PermissionRead},
		w, r,
		reqres.SecretGetUnauthorized, reqres.SecretGetBadRequest,
	)
}
