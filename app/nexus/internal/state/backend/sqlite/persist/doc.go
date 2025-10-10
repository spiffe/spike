//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package persist provides the SQLite persistence layer implementation for
// SPIKE Nexus.
//
// This package implements the backend storage using SQLite with AES-GCM
// encryption for data protection. It handles database initialization, schema
// creation, and provides thread-safe operations for storing and retrieving
// encrypted secrets and policies.
package persist
