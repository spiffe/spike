//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\ SPDX-License-Identifier: Apache-2.0

// Package backend defines the storage interface for SPIKE Nexus.
//
// This package provides the Backend interface that all storage implementations
// must satisfy. SPIKE Nexus uses this interface to persist secrets and policies
// with encryption at rest.
//
// Available implementations:
//   - sqlite: Persistent encrypted storage using SQLite (production use)
//   - memory: In-memory storage for development and testing
//   - noop: No-op implementation for embedding in other backends
//   - lite: Encryption-only backend (embeds noop, provides cipher for
//     encryption-as-a-service)
//
// The Backend interface provides:
//   - Secret storage with versioning and soft-delete support
//   - Policy storage for SPIFFE ID and path-based access control
//   - Cipher access for encryption-as-a-service endpoints
//   - Lifecycle management (Initialize/Close)
//
// All implementations must be thread-safe. Secrets and policies are encrypted
// using AES-256-GCM before storage.
package backend
