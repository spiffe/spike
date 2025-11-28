//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package ddl contains SQL statements for the SQLite backend.
//
// Despite its name, this package contains both DDL (Data Definition Language)
// and DML (Data Manipulation Language) statements:
//
//   - DDL: Schema initialization including table structures and indexes for
//     policies, secrets, and secret metadata.
//   - DML: Prepared query statements for CRUD operations on policies and
//     secrets (insert, update, select, delete).
//
// # Database Schema
//
// The schema consists of three tables:
//
// ## policies
//
// Stores access control policies with encrypted pattern matching fields.
//
//	id                         TEXT PRIMARY KEY
//	name                       TEXT NOT NULL
//	nonce                      BLOB NOT NULL
//	encrypted_spiffe_id_pattern BLOB NOT NULL
//	encrypted_path_pattern     BLOB NOT NULL
//	encrypted_permissions      BLOB NOT NULL
//	created_time               INTEGER NOT NULL
//	updated_time               INTEGER NOT NULL
//
// Note: The policies table has no additional indexes beyond the PRIMARY KEY
// because current queries only use the 'id' field (already indexed) or
// perform full table scans.
//
// ## secrets
//
// Stores versioned secret data with encryption.
//
//	path          TEXT NOT NULL
//	version       INTEGER NOT NULL
//	nonce         BLOB NOT NULL
//	encrypted_data BLOB NOT NULL
//	created_time  DATETIME NOT NULL
//	deleted_time  DATETIME (nullable, for soft deletes)
//	PRIMARY KEY (path, version)
//
// Indexes:
//   - idx_secrets_path: Index on path for efficient lookups
//   - idx_secrets_created_time: Index on created_time for time-based queries
//
// ## secret_metadata
//
// Tracks metadata for each secret path including version information.
//
//	path            TEXT PRIMARY KEY
//	current_version INTEGER NOT NULL
//	oldest_version  INTEGER NOT NULL
//	created_time    DATETIME NOT NULL
//	updated_time    DATETIME NOT NULL
//	max_versions    INTEGER NOT NULL
//
// # Query Constants
//
// The package exports the following query constants:
//
//   - QueryInitialize: Schema creation (DDL)
//   - QueryUpsertSecret: Insert or update a secret version
//   - QueryUpdateSecretMetadata: Insert or update secret metadata
//   - QuerySecretMetadata: Fetch metadata by path
//   - QuerySecretVersions: Fetch all versions of a secret
//   - QueryPathsFromMetadata: List all secret paths
//   - QueryUpsertPolicy: Insert or update a policy
//   - QueryDeletePolicy: Delete a policy by ID
//   - QueryLoadPolicy: Fetch a policy by ID
//   - QueryAllPolicies: Fetch all policies
package ddl
