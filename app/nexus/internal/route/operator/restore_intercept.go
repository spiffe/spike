//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/auth"

	"github.com/spiffe/spike/internal/net"
)

// It's unlikely to have 1000 SPIKE Keepers across the board.
// The indexes start from 1 and increase one-by-one by design.
const maxShardID = 1000

// guardRestoreRequest validates a system restore request by performing
// authentication, authorization, and input validation checks.
//
// This function implements strict authorization and validation for system
// restore operations, which are critical administrative functions that restore
// the system state from Shamir secret shares.
//
// The function performs the following validations in order:
//   - Extracts and validates the peer SPIFFE ID from the request
//   - Verifies the peer has a pilot-restore SPIFFE ID (operator role)
//   - Validates the shard ID is within valid range (1-1000)
//   - Validates the shard data is not all zeros (must contain meaningful data)
//
// Only identities with the pilot-restore role are authorized to perform system
// restore operations. The shard ID range reflects the practical limit of SPIKE
// Keeper instances in a deployment.
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The restore request containing shard ID and shard data
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the peer SPIFFE ID
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authentication fails or peer is not
//     pilot-restore
//   - apiErr.ErrInvalidInput if shard ID is out of range or shard data is
//     invalid
func guardRestoreRequest(
	request reqres.RestoreRequest, w http.ResponseWriter, r *http.Request,
) error {
	const fName = "guardRestoreRequest"

	peerSPIFFEID, err := auth.ExtractPeerSPIFFEID[reqres.ShardGetResponse](
		r, w, reqres.ShardGetUnauthorized,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	if !spiffeid.IsPilotRestore(peerSPIFFEID.String()) {
		return net.Fail(
			reqres.RestoreUnauthorized, w,
			http.StatusUnauthorized, sdkErrors.ErrUnauthorized, fName,
		)
	}

	if request.ID < 1 || request.ID > maxShardID {
		return net.Fail(
			reqres.RestoreBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	allZero := true
	for _, b := range request.Shard {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return net.Fail(
			reqres.RestoreBadInput, w,
			http.StatusBadRequest, sdkErrors.ErrInvalidInput, fName,
		)
	}

	return nil
}
