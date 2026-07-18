//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package integration exercises the SPIKE Nexus state layer end to end
// against the real sqlite persistence stack: the root-key lifecycle,
// secret and policy operations, and the operator shard-restore round
// trip (an in-process mirror of the live recovery drill under
// hack/bare-metal/drill). The package contains no production code; it
// exists so the seams between state, persist, and recovery stay covered
// by the normal test suite. See specs/integration-tests.md, Slice A.
package integration
