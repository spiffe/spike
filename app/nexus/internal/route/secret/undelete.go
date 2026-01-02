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

// RouteUndeleteSecret handles HTTP requests to restore previously deleted
// secrets.
//
// This endpoint requires authentication via SPIFFE ID and undelete permission
// for the specified secret path. It accepts a POST request with a JSON body
// containing a path to the secret and optionally specific versions to undelete.
// If no versions are specified, an empty version list is used.
//
// The function validates the request, processes the undelete operation, and
// returns a "200 OK" response upon success.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *journal.AuditEntry for logging audit information
//
// Returns:
//   - nil if the secret is successfully undeleted
//   - sdkErrors.ErrAPIPostFailed if the undelete operation fails
//   - SDK errors from request parsing or validation
//
// Request body format:
//
//	{
//	    "path": string,   // Path to the secret to undelete
//	    "versions": []int // Optional list of specific versions to undelete
//	}
//
// Responses:
//   - 200 OK: Secret successfully undeleted
//   - 400 Bad Request: Invalid request body or parameters
//   - 401 Unauthorized: Missing SPIFFE ID or insufficient permissions
//   - 500 Internal Server Error: Database operation failure
//
// The function logs its progress at various stages using structured logging.
func RouteUndeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	const fName = "routeUndeleteSecret"

	journal.AuditRequest(fName, r, audit, journal.AuditUndelete)

	request, err := net.ReadParseAndGuard[
		reqres.SecretUndeleteRequest, reqres.SecretUndeleteResponse,
	](
		w, r, reqres.SecretUndeleteResponse{}.BadRequest(),
		guardSecretUndeleteRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	undeleteErr := state.UndeleteSecret(path, versions)
	if undeleteErr != nil {
		return net.RespondWithHTTPError(
			undeleteErr, w, reqres.SecretUndeleteResponse{},
		)
	}

	return net.Success(reqres.SecretUndeleteResponse{}.Success(), w)
}
