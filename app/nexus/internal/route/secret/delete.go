//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteDeleteSecret handles HTTP DELETE requests for secret deletion
// operations. It validates the JWT token, processes the deletion request,
// and manages the secret deletion workflow.
//
// The function expects a request body containing a path and optional version
// numbers of the secrets to be deleted. If no versions are specified, an empty
// slice is used.
//
// Parameters:
//   - w: http.ResponseWriter for writing the HTTP response
//   - r: *http.Request containing the incoming HTTP request details
//   - audit: *journal.AuditEntry for logging audit information about the deletion
//     operation
//
// Returns:
//   - error: Returns nil on successful execution, or an error describing what
//     went wrong
//
// The function performs the following steps:
//  1. Validates the JWT token against the admin token
//  2. Reads and parses the request body
//  3. Processes the secret deletion
//  4. Returns a JSON response
//
// Example request body:
//
//	{
//	    "path": "secret/path",
//	    "versions": [1, 2, 3]
//	}
//
// Possible errors:
//   - "invalid or missing JWT token": When JWT validation fails
//   - "failed to read request body": When request body cannot be read
//   - "failed to parse request body": When request body is invalid
//   - "failed to marshal response body": When response cannot be serialized
func RouteDeleteSecret(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeDeleteSecret"
	journal.AuditRequest(fName, r, audit, journal.AuditDelete)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretDeleteRequest, reqres.SecretDeleteResponse](
		requestBody, w,
		reqres.SecretDeleteResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return errors.ErrParseFailure
	}

	err := guardDeleteSecretRequest(*request, w, r)
	if err != nil {
		return err
	}

	path := request.Path

	versions := request.Versions
	if len(versions) == 0 {
		versions = []int{}
	}

	err = state.DeleteSecret(path, versions)
	if err != nil {
		log.Log().Error(fName, "message", "Failed to delete secret", "err", err)
	} else {
		log.Log().Info(fName, "message", "Secret deleted")
	}

	responseBody := net.MarshalBody(reqres.SecretDeleteResponse{}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
