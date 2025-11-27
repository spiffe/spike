//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
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
// ResponseWriter, and an error is returned.
//
// Parameters:
//   - request: The secret put request containing the path and values
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - sdkErrors.ErrAccessUnauthorized if authorization fails
//   - sdkErrors.ErrAPIBadRequest if path or key name validation fails
//   - SDK errors from authentication if peer SPIFFE ID extraction fails
func guardSecretPutRequest(
	request reqres.SecretPutRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.SecretPutResponse](
		r, w, reqres.SecretPutResponse{}.Unauthorized(),
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path

	pathErr := validation.ValidatePath(path)
	if invalidPath := pathErr != nil; invalidPath {
		net.Fail(
			reqres.SecretPutResponse{}.BadRequest(), w,
			http.StatusBadRequest,
		)
		pathErr.Msg = "invalid secret path: " + path
		return pathErr
	}

	values := request.Values
	for k := range values {
		nameErr := validation.ValidateName(k)
		if nameErr != nil {
			net.Fail(
				reqres.SecretPutResponse{}.BadRequest(), w,
				http.StatusBadRequest,
			)
			nameErr.Msg = "invalid key name: " + k
			return nameErr
		}
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), path,
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		net.Fail(
			reqres.SecretPutResponse{}.Unauthorized(), w,
			http.StatusUnauthorized,
		)
		failErr := *sdkErrors.ErrAccessUnauthorized.Clone()
		failErr.Msg = "unauthorized to write secret: " + path
		return &failErr
	}

	return nil
}
