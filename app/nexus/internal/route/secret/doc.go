//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package secret provides HTTP route handlers for secret management operations
// in SPIKE Nexus.
//
// This package implements the core secret storage API, providing endpoints for:
//   - List: Enumerate all secret paths accessible to the caller
//   - Get: Retrieve secret data for a specific path and version
//   - Put: Create or update secrets (upsert semantics with versioning)
//   - Delete: Soft-delete specific versions (data retained, marked deleted)
//   - Undelete: Restore soft-deleted versions
//
// # Versioning
//
// Secrets support automatic versioning. Each Put operation creates a new
// version, preserving previous versions for audit and rollback purposes.
// Version numbers start at 1 and increment with each update. Version 0
// in API requests refers to the current (latest) version.
//
// # Soft-Delete Semantics
//
// Delete operations perform soft-deletes by marking versions as deleted
// rather than removing data. This allows recovery via Undelete. The
// CurrentVersion metadata field tracks the highest non-deleted version;
// when all versions are deleted, CurrentVersion becomes 0.
//
// # Authentication and Authorization
//
// All endpoints require mTLS authentication via SPIFFE. The caller's SPIFFE ID
// is extracted from the client certificate and validated against configured
// policies. Each route handler uses guard functions to verify the caller has
// appropriate permissions (read, write, list) for the requested path.
//
// # Request Handling
//
// Route handlers follow the standard SPIKE pattern using net.ReadParseAndGuard
// for combined request reading, JSON parsing, and authorization in a single
// operation. Guard functions validate SPIFFE IDs against policies before
// allowing access to secret data.
package secret
