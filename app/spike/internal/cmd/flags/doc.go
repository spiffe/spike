//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package flags reads typed values from Cobra command flags.
//
// A flag that cannot be read was either never registered or was registered
// with a different type. That is a programming error rather than user
// input, so instead of returning a zero value that the command would then
// act on silently, these helpers return an error for the command to
// propagate; the CLI then exits non-zero.
package flags
