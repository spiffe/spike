//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package state manages the cryptographic state for SPIKE Bootstrap.
//
// This package handles the generation and management of the root key used to
// encrypt secrets, as well as the creation and distribution of Shamir secret
// shares to SPIKE Keeper instances.
package state
