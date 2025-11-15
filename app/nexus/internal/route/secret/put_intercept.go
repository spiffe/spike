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
	"github.com/spiffe/spike/internal/auth"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

// guardSecretPutRequest validates a secret storage request by performing
// authentication, authorization, and input validation checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Validates the secret path format
//   - Validates each key name in the secret values map
//   - Checks if the peer has write permission for the specified secret path
//
// Write permission is required to create or update secret data. The key name
// validation ensures that all keys in the secret values conform to naming
// requirements. The authorization check is performed against the specific
// secret path to enable fine-grained access control.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The secret put request containing the path and values
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
//   - apiErr.ErrInvalidInput if path or key name validation fails
func guardSecretPutRequest(
	request reqres.SecretPutRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.SecretPutResponse](
		r, w, reqres.SecretPutUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path

	err = validation.ValidatePath(path)
	if invalidPath := err != nil; invalidPath {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretPutBadInput, w,
		)
		if alreadyResponded := err != nil; !alreadyResponded {
			net.Respond(http.StatusBadRequest, responseBody, w)
		}
		return apiErr.ErrInvalidInput
	}

	values := request.Values
	for k := range values {
		err := validation.ValidateName(k)
		if err != nil {
			responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
				reqres.SecretPutBadInput, w,
			)
			if alreadyResponded := err != nil; !alreadyResponded {
				net.Respond(http.StatusUnauthorized, responseBody, w)
			}
			return apiErr.ErrInvalidInput
		}
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.SecretPutUnauthorized, w,
		)
		if respondedAlready := err != nil; !respondedAlready {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}
	return nil
}
