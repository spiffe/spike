//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
)

// RoutePutPolicy handles HTTP requests for creating or updating policies.
// It processes the request body to upsert a policy with the specified name,
// SPIFFE ID pattern, path pattern, and permissions.
//
// This handler follows upsert semantics consistent with secret operations:
//   - If no policy with the given name exists, a new policy is created
//   - If a policy with the same name exists, it is updated
//
// The function expects a JSON request body containing:
//   - Name: policy name (used as the unique identifier for upsert)
//   - SPIFFEIDPattern: SPIFFE ID matching pattern (regex)
//   - PathPattern: path matching pattern (regex)
//   - Permissions: set of allowed permissions
//
// On success, it returns a JSON response with the policy's ID.
// On failure, it returns an appropriate error response with status code.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy data
//   - audit: Audit entry for logging the policy upsert action
//
// Returns:
//   - *sdkErrors.SDKError: nil on successful policy upsert, error otherwise
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
) *sdkErrors.SDKError {
	const fName = "RoutePutPolicy"

	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	request, err := net.ReadParseAndGuard[
		reqres.PolicyPutRequest, reqres.PolicyPutResponse,
	](
		w, r, reqres.PolicyPutResponse{}.BadRequest(), guardPolicyCreateRequest,
	)
	if alreadyResponded := err != nil; alreadyResponded {
		return err
	}

	policy, upsertErr := state.UpsertPolicy(data.Policy{
		Name:            request.Name,
		SPIFFEIDPattern: request.SPIFFEIDPattern,
		PathPattern:     request.PathPattern,
		Permissions:     request.Permissions,
	})
	if upsertErr != nil {
		return net.RespondWithHTTPError(upsertErr, w, reqres.PolicyPutResponse{})
	}

	return net.Success(reqres.PolicyPutResponse{ID: policy.ID}.Success(), w)
}
