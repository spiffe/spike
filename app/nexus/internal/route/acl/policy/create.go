//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	stdErrs "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RoutePutPolicy handles HTTP PUT requests for creating new policies.
// It processes the request body to create a policy with the specified name,
// SPIFFE ID pattern, path pattern, and permissions.
//
// The function expects a JSON request body containing:
//   - Name: policy name
//   - SpiffeIdPattern: SPIFFE ID matching pattern
//   - PathPattern: path matching pattern
//   - Permissions: set of allowed permissions
//
// On success, it returns a JSON response with the created policy's ID.
// On failure, it returns an appropriate error response with status code.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy creation data
//   - audit: Audit entry for logging the policy creation action
//
// Returns:
//   - error: nil on successful policy creation, error otherwise
//
// Example request body:
//
//	{
//	    "name": "example-policy",
//	    "spiffe_id_pattern": "^spiffe://example\\.org/.*/service$",
//	    "path_pattern": "^secrets/db/.*$",
//	    "permissions": ["read", "write"]
//	}
//
// Example success response:
//
//	{
//	    "id": "policy-123"
//	}
//
// Example error response:
//
//	{
//	    "err": "Internal server error"
//	}
func RoutePutPolicy(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "RoutePutPolicy"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyCreateRequest, reqres.PolicyCreateResponse,
	](
		w, r, reqres.PolicyCreateBadInput, guardPolicyCreateRequest, fName,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		log.Log().Error(fName, "message", "exit", "err", err.Error())
		return err
	}

	name := request.Name
	SPIFFEIDPattern := request.SPIFFEIDPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	policy, err := state.CreatePolicy(data.Policy{
		Name:            name,
		SPIFFEIDPattern: SPIFFEIDPattern,
		PathPattern:     pathPattern,
		Permissions:     permissions,
	})
	if err != nil {
		failErr := stdErrs.Join(errors.ErrCreationFailed, err)
		return net.Fail(
			reqres.PolicyCreateInternal, w,
			http.StatusInternalServerError, failErr, fName,
		)
	}

	net.Success(reqres.PolicyCreateResponse{ID: policy.ID}.Success(), w, fName)
	return nil
}
