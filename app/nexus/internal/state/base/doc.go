//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package base provides the high-level state management API for SPIKE Nexus.
//
// This package is the primary interface used by route handlers to manage
// secrets and policies. It wraps the lower-level persist and backend packages,
// providing business logic such as versioning, access control, and validation.
//
// Key functions:
//
// Initialization:
//   - Initialize: Sets up the backend storage with a root encryption key
//
// Secret management:
//   - UpsertSecret: Store or update a secret with automatic versioning
//   - GetSecret: Retrieve a secret by path and optional version
//   - DeleteSecret: Soft-delete specific versions of a secret
//   - UndeleteSecret: Restore soft-deleted versions
//   - ListSecrets: List all secret paths
//
// Policy management:
//   - UpsertPolicy: Create or update an access control policy
//   - GetPolicy: Retrieve a policy by ID
//   - DeletePolicy: Remove a policy
//   - ListPolicies: List policies with optional filtering
//   - CheckAccess: Evaluate if a SPIFFE ID has required permissions for a path
//
// Access control:
//
// SPIKE Pilot instances have unrestricted access. Other workloads must have
// matching policies. A policy matches when its SPIFFE ID pattern and path
// pattern both match the request, and it grants the required permissions.
//
// This package delegates storage operations to the persist package, which in
// turn uses the configured backend (sqlite, memory, lite, or noop).
package base
