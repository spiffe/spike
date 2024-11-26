//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

// AdminTokenWriteRequest is to persist the admin token in memory.
// Admin token can be persisted only once. It is used to receive a
// short-lived session token.
type AdminTokenWriteRequest struct {
	Data string `json:"data"`
}

// AdminTokenWriteResponse is to persist the admin token in memory.
type AdminTokenWriteResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}
