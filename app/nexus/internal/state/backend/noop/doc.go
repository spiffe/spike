//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package noop provides a no-operation storage backend for SPIKE Nexus.
//
// This backend implements the backend.Backend interface but performs no actual
// storage or retrieval operations. All methods are no-ops that return nil or
// empty values without side effects.
//
// Use this backend when:
//   - Testing components that depend on a backend without actual storage
//   - Embedding in other backends to provide default no-op behavior (e.g., lite)
//   - Placeholder during development before a real backend is configured
//
// Key characteristics:
//   - No storage: All Store operations are silently ignored
//   - No retrieval: All Load operations return nil
//   - No errors: All operations succeed (return nil error)
//   - No cipher: GetCipher returns nil (no encryption capability)
//   - Thread-safe: No state means no synchronization needed
//
// The lite backend embeds this Store to inherit no-op storage behavior while
// providing its own cipher implementation for encryption-as-a-service.
//
// For actual storage, use the memory or sqlite backends instead.
package noop
