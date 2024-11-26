//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

type ErrorCode string

var ErrBadInput = ErrorCode("bad_request")
var ErrServerFault = ErrorCode("server_fault")
var ErrUnauthorized = ErrorCode("unauthorized")
var ErrLowEntropy = ErrorCode("low_entropy")
var ErrAlreadyInitialized = ErrorCode("already_initialized")
var ErrNotFound = ErrorCode("not_found")
