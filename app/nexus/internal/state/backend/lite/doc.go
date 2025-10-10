//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lite provides an encryption-only backend implementation for SPIKE
// Nexus.
//
// This backend provides encryption-as-a-service functionality without
// persisting any data to a backing store. It is intended for scenarios where
// secrets are stored externally in S3-compatible storage, and SPIKE Nexus
// only provides cryptographic operations. In this mode, SPIKE policies are
// minimally enforced.
package lite
