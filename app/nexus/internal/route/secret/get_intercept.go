//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
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
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if path validation fails
func guardGetSecretRequest(
	request reqres.SecretReadRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	path := request.Path
	err = validation.ValidatePath(path)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretReadBadInput, w,
		)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(),
		path,
		[]data.PolicyPermission{data.PermissionRead},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretReadUnauthorized, w,
		)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	return nil
}
