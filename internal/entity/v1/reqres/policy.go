//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

type CreatePolicyRequest struct {
	Name            string   `json:"name"`
	SpiffeIdPattern string   `json:"spiffe_id_pattern"`
	PathPattern     string   `json:"path_pattern"`
	Permissions     []string `json:"permissions"`
}

type CreatePolicyResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}

type CheckAccessRequest struct {
	SpiffeID string `json:"spiffe_id"`
	Path     string `json:"path"`
	Action   string `json:"action"`
}

type CheckAccessResponse struct {
	Allowed          bool     `json:"allowed"`
	MatchingPolicies []string `json:"matching_policies"`
}
