//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package entity

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrParseFailure = errors.New("failed to parse request body")
var ErrReadFailure = errors.New("failed to read request body")
var ErrMarshalFailure = errors.New("failed to marshal response body")
var ErrAlreadyInitialized = errors.New("already initialized")
