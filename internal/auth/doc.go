//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package auth provides utility functions for identity verification and
// authorization using SPIFFE IDs within the SPIKE system.
//
// This package ensures secure communication and proper role identification
// between different SPIKE components (e.g., Pilot, Keeper, and Nexus). It
// provides validation and access control mechanisms by matching provided
// SPIFFE IDs against pre-defined values.
package auth
