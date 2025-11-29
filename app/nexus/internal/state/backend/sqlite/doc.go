//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package sqlite provides a persistent, encrypted SQLite storage backend for
// SPIKE Nexus.
//
// This backend stores secrets and policies in a SQLite database with
// AES-256-GCM encryption. All sensitive data is encrypted at rest, providing
// defense-in-depth for secret storage.
//
// Use this backend when:
//   - Persistent storage across restarts is required
//   - Secrets must be encrypted at rest
//   - A single-node deployment is sufficient (SQLite is not distributed)
//
// Key characteristics:
//   - Persistent: Data survives process restarts
//   - Encrypted: All secrets encrypted with AES-256-GCM
//   - Versioned: Supports configurable secret version retention
//   - Full policy support: Stores and enforces access control policies
//   - Single-node: SQLite does not support distributed deployments
//
// The database is stored at ~/.spike/data/spike.db by default. The encryption
// key must be exactly 32 bytes (AES-256).
//
// For non-persistent scenarios, use the memory backend. For encryption-only
// without local storage, use the lite backend.
//
// The actual persistence implementation is in the `persist` subpackage.
package sqlite

import (
	// Imported for side effects to register the SQLite driver.
	_ "github.com/mattn/go-sqlite3"
)
