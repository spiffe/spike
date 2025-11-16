//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// guardListSecretRequest validates a secret listing request by performing
// authentication and authorization checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer has list permission for the system secret access path
//
// List permission is required to enumerate secrets in the system. The
// authorization check is performed against the system-level secret access path
// to control which identities can discover what secrets exist.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The secret list request (currently unused, reserved for future
//     validation needs)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
func guardListSecretRequest(
	_ reqres.SecretListRequest, w http.ResponseWriter, r *http.Request,
) error {
	const fName = "guardListSecretRequest"

	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.SecretListResponse](
		r, w, reqres.SecretListUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	allowed := state.CheckAccess(
		peerSPIFFEID.String(), apiAuth.PathSystemSecretAccess,
		[]data.PolicyPermission{data.PermissionList},
	)
	if !allowed {
		return net.Fail(
			reqres.SecretListResponse{Err: data.ErrUnauthorized}, w,
			http.StatusUnauthorized, apiErr.ErrUnauthorized, fName,
		)
	}

	return nil
}
