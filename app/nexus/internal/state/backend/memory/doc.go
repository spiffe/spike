//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package memory provides a fully functional in-memory storage backend for
// SPIKE Nexus.
//
// This backend stores secrets and policies in memory using thread-safe data
// structures. It provides complete secret management functionality including
// versioning, metadata tracking, and policy enforcement.
//
// Use this backend when:
//   - Running in development or testing environments
//   - Persistent storage is not required
//   - Fast access without disk I/O is preferred
//   - Data loss on restart is acceptable
//
// Key characteristics:
//   - Thread-safe: Uses separate RWMutex for secrets and policies
//   - Versioned secrets: Supports configurable max versions per secret
//   - Full policy support: Stores and enforces access control policies
//   - No persistence: All data is lost when the process terminates
//
// For persistent storage, use the sqlite backend instead. For encryption-only
// scenarios without local storage, use the lite backend.
package memory
