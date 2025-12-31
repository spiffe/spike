//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	cfg "github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

func hasListPermission(peerSPIFFEID string) bool {
	return state.CheckAccess(
		peerSPIFFEID, cfg.PathSystemPolicyAccess,
		[]data.PolicyPermission{data.PermissionList},
	)
}

// guardListPolicyRequest validates a policy list request by performing
// authentication and authorization checks.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer has list permission for the policy access path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The policy list request (currently unused, reserved for future
//     validation needs)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
func guardListPolicyRequest(
	_ reqres.PolicyListRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	return net.RespondUnauthorizedOnPredicateFail(hasListPermission,
		reqres.PolicyListResponse{}.Unauthorized(), w, r)
}
