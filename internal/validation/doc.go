//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package validation provides runtime validation helpers for SPIKE components.
//
// This package contains functions that validate preconditions and invariants.
// When validation fails, the functions terminate the program rather than
// returning errors, as the conditions they check represent programming errors
// that should never occur in production.
//
// Current validations:
//
//   - CheckContext: Ensures a context.Context is not nil. A nil context
//     indicates a programming error in the caller.
//
// Usage:
//
//	func SomeOperation(ctx context.Context) error {
//	    validation.CheckContext(ctx, "SomeOperation")
//	    // ... proceed knowing ctx is valid
//	}
//
// This package is internal to SPIKE and should not be imported by external
// code.
package validation
