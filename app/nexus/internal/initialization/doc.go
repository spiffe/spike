//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package initialization handles SPIKE Nexus startup based on the configured
// backend store type.
//
// # Backend Types
//
// The package supports three backend configurations:
//
//   - sqlite: Production mode with SQLite persistence. Recovers the root key
//     from SPIKE Keepers and starts periodic shard synchronization.
//   - lite: Lightweight mode with SQLite persistence but without root key
//     encryption. Still uses Keepers for consistency.
//   - memory: Development mode with in-memory storage. No Keepers required,
//     data is lost on restart. Not for production use.
//
// # Initialization Flow
//
// For sqlite and lite backends:
//  1. Contact SPIKE Keeper instances to collect root key shards
//  2. Reconstruct the root key using Shamir's Secret Sharing
//  3. Initialize the backing store with the recovered key
//  4. Start a background goroutine for periodic shard distribution
//
// For memory backend:
//  1. Initialize an empty in-memory store without the root key
//  2. Log warnings about non-production use
//
// # Subpackages
//
// The recovery subpackage implements the actual shard collection, root key
// reconstruction, and Keeper communication logic.
package initialization
