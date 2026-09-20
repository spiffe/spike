//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package layout holds the repository-wide source layout rules that are
// enforced by tests rather than by a linter.
//
// The package has no runtime code. Its tests parse every Go file in the
// module and fail when a file breaks a house rule, in the same way that the
// route packages' guard tests fail when a handler skips its guard. The
// first rule: a file holds either exported or unexported functions and
// methods, never both, so that the public surface of a package stays in a
// small file and implementation detail lives in a semantically named
// sibling (for example error.go and error_impl.go), with test helpers in
// test_helper_test.go.
package layout
