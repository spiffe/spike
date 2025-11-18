//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	stdErrs "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteUndeleteSecret handles HTTP requests to restore previously deleted
// secrets.
//
// This endpoint requires a valid admin JWT token for authentication. It accepts
// a POST request with a JSON body containing a path to the secret and
// optionally specific versions to undelete. If no versions are specified,
// an empty version list is used.
//
// The function validates the JWT, reads and unmarshals the request body,
// processes the undelete operation, and returns a "200 OK" response upon
// success.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *journal.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
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
//   - 401 Unauthorized: Invalid or missing JWT token
//
// The function logs its progress at various stages using structured logging.
func RouteUndeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeUndeleteSecret"

	journal.AuditRequest(fName, r, audit, journal.AuditUndelete)

	request, err := net.ReadParseAndGuard[
		reqres.SecretUndeleteRequest, reqres.SecretUndeleteResponse,
	](
		w, r, reqres.SecretUndeleteBadInput, guardSecretUndeleteRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	path := request.Path
	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	err = state.UndeleteSecret(path, versions)
	if err != nil {
		failErr := stdErrs.Join(sdkErrors.ErrUndeleteFailed, err)
		return net.Fail(
			reqres.SecretUndeleteInternal, w,
			http.StatusInternalServerError, failErr, fName,
		)
	}

	net.Success(reqres.SecretUndeleteSuccess.Success(), w, fName)
	return nil
}
