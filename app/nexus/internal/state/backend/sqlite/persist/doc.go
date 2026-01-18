//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package persist provides the SQLite persistence layer implementation for
// SPIKE Nexus.
//
// This package implements the backend.Backend interface using SQLite with
// AES-256-GCM encryption for data protection. It is the internal implementation
// used by the parent sqlite package.
//
// Key components:
//   - DataStore: Main struct implementing backend.Backend interface
//   - Schema management: Automatic table creation and migrations
//   - Encryption: AES-GCM encryption/decryption for secrets at rest
//   - Nonce generation: Secure random nonce generation for each encryption
//
// Thread safety:
//   - Uses sync.RWMutex for concurrent read/write access
//   - Uses sync.Once for safe database closure
//
// File organization:
//   - ds.go: DataStore struct definition
//   - initialize.go: Database initialization and schema setup
//   - secret.go, secret_load.go: Secret storage and retrieval
//   - policy.go: Policy storage and retrieval
//   - transform.go: Data serialization/deserialization
//   - schema.go: Database schema definitions
//   - options.go, parse.go: Configuration parsing
//
// This package is not intended to be used directly. Use the parent sqlite
// package instead, which provides the public New() constructor.
package persist
