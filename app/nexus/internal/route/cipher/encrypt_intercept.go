//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

//func spiffeidAllowedForEncryptCipher(spiffeid string) bool {
//	// Lite Workloads are always allowed:
//	allowed := false
//	if sdkSpiffeid.IsLiteWorkload(spiffeid) {
//		allowed = true
//	}
//	// If not, do a policy check to determine if the request is allowed:
//	if !allowed {
//		allowed = state.CheckAccess(
//			spiffeid,
//			apiAuth.PathSystemCipherExecute,
//			[]data.PolicyPermission{data.PermissionExecute},
//		)
//	}
//	return allowed
//}

// guardCipherEncryptRequest validates a cipher encryption request by
// performing authentication, authorization, and request field validation.
//
// This function implements a two-tier authorization model:
//  1. Lite workloads are automatically granted encryption access
//  2. Other workloads must have execute permission for the cipher encrypt path
//
// The function performs the following validations in order:
//   - Validates request fields (future: size limits, format checks, etc.)
//   - Checks if the peer is a lite workload (automatically allowed)
//   - If not a lite workload, checks if the peer has execute permission for
//     the system cipher encrypt path
//
// If any validation fails, an appropriate error response is written to the
// ResponseWriter and an error is returned.
//
// Parameters:
//   - request: The cipher encryption request to validate
//   - peerSPIFFEID: The already-validated peer SPIFFE ID (pointer)
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request (for context)
//
// Returns:
//   - nil if all validations pass
//   - apiErr.ErrUnauthorized if authorization fails
//   - apiErr.ErrBadInput if request validation fails
func guardCipherEncryptRequest(
	request reqres.CipherEncryptRequest,
	w http.ResponseWriter,
	r *http.Request,
) *sdkErrors.SDKError {
	if authErr := net.AuthorizeAndRespondOnFail(
		reqres.CipherEncryptResponse{}.Unauthorized(),
		predicate.AllowSPIFFEIDForCipherEncrypt,
		state.CheckAccess,
		w, r,
	); authErr != nil {
		return authErr
	}

	// Validate plaintext size to prevent DoS attacks
	return net.RespondCryptoErrOnLargeCipherText(
		request.Plaintext, w, reqres.CipherEncryptResponse{}.BadRequest(),
	)
}
