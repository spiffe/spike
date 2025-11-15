//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	apiAuth "github.com/spiffe/spike-sdk-go/config/auth"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/net"
)

// guardDecryptCipherRequest validates a cipher decryption request by
// performing authentication and authorization checks.
//
// This function implements a two-tier authorization model:
//  1. Lite workloads are automatically granted decryption access
//  2. Other workloads must have execute permission for the cipher decrypt path
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Checks if the peer is a lite workload (automatically allowed)
//   - If not a lite workload, checks if the peer has execute permission for
//     the system cipher decrypt path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The cipher decryption request (currently unused, reserved for
//     future validation needs)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication or authorization fails
func guardDecryptCipherRequest(
	_ reqres.CipherDecryptRequest, w http.ResponseWriter, r *http.Request,
) error {
	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetResponse{
			Err: data.ErrUnauthorized,
		})
	alreadyResponded := err != nil
	if alreadyResponded {
		return err
	}

	// Lite workloads are always allowed:
	allowed := false
	if spiffeid.IsLiteWorkload(peerSPIFFEID.String()) {
		allowed = true
	}
	// If not, do a policy check to determine if the request is allowed:
	if !allowed {
		allowed = state.CheckAccess(
			peerSPIFFEID.String(),
			apiAuth.PathSystemCipherDecrypt,
			[]data.PolicyPermission{data.PermissionExecute},
		)
	}

	if !allowed {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			reqres.CipherDecryptResponse{
				Err: data.ErrUnauthorized,
			}, w)
		alreadyResponded = err != nil
		if !alreadyResponded {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return apiErr.ErrUnauthorized
	}

	return nil
}
