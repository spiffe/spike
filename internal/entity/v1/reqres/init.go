//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

import "github.com/spiffe/spike/internal/entity/data"

// CheckInitStateRequest is to check if the SPIKE Keep is initialized.
type CheckInitStateRequest struct {
}

// CheckInitStateResponse is to check if the SPIKE Keep is initialized.
type CheckInitStateResponse struct {
	State data.InitState `json:"state"`
	Err   ErrorCode      `json:"err,omitempty"`
}

// InitRequest is to initialize SPIKE as a superuser.
type InitRequest struct {
	// Password string `json:"password"`
}

// InitResponse is to initialize SPIKE as a superuser.
type InitResponse struct {
	RecoveryToken string    `json:"token"`
	Err           ErrorCode `json:"err,omitempty"`
}
