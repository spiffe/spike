//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package integration holds SPIKE's live, opt-in Pilot integration tests
// (specs/integration-tests.md, Slice B). The tests drive the built `spike`
// binary end to end against a running `make start` environment (SPIRE,
// Nexus, and the Keepers), rather than importing command internals, so
// they exercise the same seams a real operator does.
//
// The suite is build-tagged `integration` and additionally gated at run
// time by SPIKE_INTEGRATION_TEST=1, so it never runs in the normal
// `make test`. Two cases are non-destructive (a CRUD and cipher smoke
// pass against a healthy environment, and the Pilot's connection warning
// when Nexus is unreachable). The third proves the Pilot denies work when
// Nexus is reachable but uninitialized; reaching that state means killing
// Nexus and every Keeper and restarting Nexus alone, so it is gated behind
// a second flag, SPIKE_INTEGRATION_DESTRUCTIVE=1. It leaves the environment
// uninitialized but cleans up the Nexus it spawned, so a Ctrl+C on the
// make start terminal (or `make kill`) followed by `make start` resets it.
// It complements, and does not subsume, the live recovery drill under
// hack/bare-metal/drill.
package integration
