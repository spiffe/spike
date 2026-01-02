//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike-sdk-go/journal"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// RouteListPaths handles requests to retrieve all available secret paths.
//
// This endpoint requires the peer to have list permission for the system
// secret access path. The function returns a list of all paths where secrets
// are stored, regardless of their version or deletion status.
//
// The function follows these steps:
//  1. Validates peer SPIFFE ID and authorization (via guardListSecretRequest)
//  2. Validates the request body format
//  3. Retrieves all secret paths from the state
//  4. Returns the list of paths
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request containing the peer SPIFFE ID
//   - audit: The audit entry for logging audit information
//
// Returns:
//   - *sdkErrors.SDKError: An error if validation or processing fails.
//     Returns nil on success.
//
// Request body format:
//
//	{} // Empty request body expected
//
// The response format on success (200 OK):
//
//	{
//	    "keys": []string   // Array of all secret paths
//	}
//
// Error responses:
//   - 401 Unauthorized: Authentication or authorization failure
//   - 400 Bad Request: Invalid request body format
//
// All operations are logged using structured logging. This endpoint only
// returns the paths to secrets and not their contents; use RouteGetSecret to
// retrieve actual secret values.
func RouteListPaths(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RouteListPaths"

	journal.AuditRequest(fName, r, audit, journal.AuditList)

	_, err := net.ReadParseAndGuard[
		reqres.SecretListRequest, reqres.SecretListResponse](
		w, r, reqres.SecretListResponse{}.BadRequest(), guardListSecretRequest,
	)
	if err != nil {
		return err
	}

	return net.Success(
		reqres.SecretListResponse{Keys: state.ListKeys()}.Success(), w,
	)
}
