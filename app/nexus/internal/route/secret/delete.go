//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
)

// RouteDeleteSecret handles HTTP DELETE requests for secret deletion
// operations. It authenticates the peer, validates permissions, processes
// the deletion request, and manages the secret deletion workflow.
//
// The function expects a request body containing a path and optional version
// numbers of the secrets to be deleted. If no versions are specified, the
// current version is deleted.
//
// Parameters:
//   - w: http.ResponseWriter for writing the HTTP response
//   - r: *http.Request containing the incoming HTTP request with peer SPIFFE ID
//   - audit: *journal.AuditEntry for logging audit information about the
//     deletion operation
//
// Returns:
//   - *sdkErrors.SDKError: Returns nil on successful execution, or an error
//     describing what went wrong
//
// The function performs the following steps:
//  1. Authenticates the peer via SPIFFE ID and validates write permissions
//  2. Reads and parses the request body
//  3. Processes the secret deletion (soft-delete operation)
//  4. Returns an appropriate JSON response
//
// Example request body:
//
//	{
//	    "path": "secret/path",
//	    "versions": [1, 2, 3]
//	}
//
// Response codes:
//   - 200 OK: Secret successfully deleted
//   - 400 Bad Request: Invalid request body or path format
//   - 401 Unauthorized: Authentication or authorization failure
//   - 404 Not Found: Secret does not exist at the specified path
//   - 500 Internal Server Error: Backend or server-side failure
func RouteDeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "RouteDeleteSecret"

	journal.AuditRequest(fName, r, audit, journal.AuditDelete)

	request, err := net.ReadParseAndGuard[
		reqres.SecretDeleteRequest, reqres.SecretDeleteResponse](
		w, r, reqres.SecretDeleteResponse{}.BadRequest(), guardDeleteSecretRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	deleteErr := state.DeleteSecret(path, versions)
	if deleteErr != nil {
		return net.RespondWithHTTPError(deleteErr, w, reqres.SecretDeleteResponse{})
	}

	return net.Success(reqres.SecretDeleteResponse{}.Success(), w)
}
