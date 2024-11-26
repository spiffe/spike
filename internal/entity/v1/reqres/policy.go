//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

import "github.com/spiffe/spike/internal/entity/data"

type PolicyCreateRequest struct {
	Name            string                  `json:"name"`
	SpiffeIdPattern string                  `json:"spiffe_id_pattern"`
	PathPattern     string                  `json:"path_pattern"`
	Permissions     []data.PolicyPermission `json:"permissions"`
}

type PolicyCreateResponse struct {
	Id  string    `json:"id,omitempty"`
	Err ErrorCode `json:"err,omitempty"`
}

type PolicyReadRequest struct {
	Id string `json:"id"`
}

type PolicyReadResponse struct {
	data.Policy
	Err ErrorCode `json:"err,omitempty"`
}

type PolicyDeleteRequest struct {
	Id string `json:"id"`
}

type PolicyDeleteResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}

type PolicyListRequest struct{}

type PolicyListResponse struct {
	Policies []data.Policy `json:"policies"`
	Err      ErrorCode     `json:"err,omitempty"`
}

type PolicyAccessCheckRequest struct {
	SpiffeID string `json:"spiffe_id"`
	Path     string `json:"path"`
	Action   string `json:"action"`
}

type PolicyAccessCheckResponse struct {
	Allowed          bool      `json:"allowed"`
	MatchingPolicies []string  `json:"matching_policies"`
	Err              ErrorCode `json:"err,omitempty"`
}
