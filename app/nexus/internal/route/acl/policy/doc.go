//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package policy provides HTTP route handlers for access control policy
// management in SPIKE Nexus.
//
// This package implements the policy API, providing endpoints for:
//   - List: Enumerate all defined policies
//   - Get: Retrieve a specific policy by ID
//   - Put: Create or update a policy (upsert semantics)
//   - Delete: Remove a policy
//
// # Policy Structure
//
// Policies define access control rules that map SPIFFE IDs to secret paths
// with specific permissions. Each policy contains:
//
//   - Name: Human-readable identifier for the policy
//   - SPIFFEIDPattern: Regular expression matching allowed SPIFFE IDs
//   - PathPattern: Regular expression matching allowed secret paths
//   - Permissions: List of granted permissions (read, write, list, super)
//
// The "super" permission acts as a wildcard, granting all other permissions.
//
// # Pattern Matching
//
// Both SPIFFEIDPattern and PathPattern use regular expressions (not globs).
// When a workload attempts to access a secret, its SPIFFE ID is matched
// against SPIFFEIDPattern and the requested path against PathPattern. Access
// is granted only if both patterns match and the required permission is
// present.
//
// # Upsert Semantics
//
// The Put endpoint uses upsert semantics: if a policy with the same name
// exists, it is updated (preserving ID and CreatedAt); otherwise, a new
// policy is created. This differs from create-only semantics and allows
// policies to be modified without delete/recreate cycles.
//
// # Authentication and Authorization
//
// All endpoints require mTLS authentication via SPIFFE. Policy management
// endpoints are restricted to SPIKE Pilot, which acts as the administrative
// interface for SPIKE. Guard functions validate the caller's SPIFFE ID
// before allowing policy operations.
package policy
