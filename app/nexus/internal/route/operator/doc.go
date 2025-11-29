//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package operator provides HTTP route handlers for SPIKE Nexus operator
// operations.
//
// This package implements disaster recovery endpoints that allow operators
// to recover Nexus when normal operation is not possible:
//
//   - Recover: Retrieves recovery shards that can be used to reconstruct
//     the root key. Returns Shamir secret shares that must be combined
//     to restore Nexus operation.
//
//   - Restore: Accepts recovery shards to reconstruct the root key and
//     restore Nexus to operational state. Requires the threshold number
//     of valid shards to succeed.
//
// # Security Constraints
//
// These endpoints are restricted to SPIKE Pilot only. The guard functions
// validate that the caller's SPIFFE ID matches the expected SPIKE Pilot
// identity before allowing access. This prevents unauthorized workloads
// from accessing or manipulating recovery material. See ADR-0029 for the
// rationale behind this restriction.
//
// # Authentication
//
// All endpoints require mTLS authentication via SPIFFE. The caller's SPIFFE ID
// is extracted from the client certificate and validated before any recovery
// operations are performed.
package operator
