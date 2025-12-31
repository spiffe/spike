//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffeid"
)

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
//   - *sdkErrors.SDKError: An error if authentication fails, the peer is not
//     authorized (not pilot-restore), the shard ID is out of range, or the
//     shard data is invalid. Returns nil if all validations pass.
func guardRestoreRequest(
	request reqres.RestoreRequest, w http.ResponseWriter, r *http.Request,
) *sdkErrors.SDKError {
	err := net.RespondUnauthorizedOnPredicateFail(spiffeid.IsPilotRestore,
		reqres.RestoreResponse{}.Unauthorized(), w, r,
	)
	if err != nil {
		return err
	}

	// TODO: magic number: 1
	if request.ID < 1 || request.ID > env.ShamirMaxShareCountVal() {
		failErr := net.Fail(
			reqres.RestoreResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrAPIBadRequest.Wrap(failErr)
		}
		return sdkErrors.ErrAPIBadRequest
	}

	allZero := true
	for _, b := range request.Shard {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		failErr := net.Fail(
			reqres.RestoreResponse{}.BadRequest(), w, http.StatusBadRequest,
		)
		if failErr != nil {
			return sdkErrors.ErrAPIBadRequest.Wrap(failErr)
		}
		return sdkErrors.ErrAPIBadRequest
	}
	return nil
}
