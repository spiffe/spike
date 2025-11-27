//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package persist manages the global backend instance for SPIKE Nexus.
//
// This package acts as a singleton manager that initializes and provides
// thread-safe access to the configured storage backend. It sits between the
// high-level state/base package and the low-level backend implementations.
//
// Architecture:
//
//	route handlers -> state/base -> persist -> backend implementations
//
// Key functions:
//   - InitializeBackend: Creates the backend based on SPIKE_NEXUS_BACKEND_STORE
//     environment variable (sqlite, memory, or lite)
//   - Backend: Returns the initialized backend instance (thread-safe)
//
// Backend selection (via SPIKE_NEXUS_BACKEND_STORE):
//   - "sqlite": Persistent encrypted SQLite storage (production)
//   - "memory": In-memory storage (development/testing)
//   - "lite": Encryption-only, no persistence (encryption-as-a-service)
//   - default: Falls back to memory
//
// Thread safety:
//
// Both InitializeBackend and Backend use mutex locking to ensure safe
// concurrent access. InitializeBackend should be called once during startup;
// Backend can be called from any goroutine.
package persist
