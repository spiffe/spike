//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package secret implements SPIKE CLI commands for managing secrets.
//
// Secrets in SPIKE are versioned key-value pairs stored in SPIKE Nexus with
// AES-256-GCM encryption at rest. Access is controlled by policies that match
// SPIFFE IDs and path patterns. All operations use SPIFFE-based authentication.
//
// Available commands:
//
//   - put: Store a secret with one or more key-value pairs at a path. Creates
//     a new version if the secret already exists.
//   - get: Retrieve a secret's values. Supports fetching specific versions.
//   - list: List secret paths, optionally filtered by a path prefix.
//   - delete: Soft-delete secret versions. Deleted versions can be recovered.
//   - undelete: Restore previously soft-deleted versions.
//   - metadata-get: Retrieve secret metadata (versions, timestamps) without
//     the actual values.
//
// Secret paths:
//
// Paths are namespace identifiers, not filesystem paths. They should not start
// with a forward slash:
//
//	spike secret put secrets/db/password ...  # correct
//	spike secret put /secrets/db/password ... # incorrect
//
// Versioning:
//
// Each put operation creates a new version. Version 0 always refers to the
// current (latest) version. Older versions can be retrieved, deleted, or
// restored individually.
//
// Example usage:
//
//	spike secret put secrets/db/creds user=admin pass=secret
//	spike secret get secrets/db/creds
//	spike secret get secrets/db/creds --version 1
//	spike secret list secrets/db
//	spike secret delete secrets/db/creds --versions 1,2
//	spike secret undelete secrets/db/creds --versions 1,2
//	spike secret metadata-get secrets/db/creds
//
// See https://spike.ist/usage/commands/ for detailed CLI documentation.
package secret
