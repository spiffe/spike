//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package state manages the Shamir secret share for SPIKE Keeper.
//
// This package provides thread-safe storage and access to a single shard of
// the root key. SPIKE Nexus distributes shards to multiple Keeper instances
// using Shamir's Secret Sharing scheme. When Nexus needs to recover its root
// key, it collects shards from the required number of Keepers and reconstructs
// the original secret.
//
// The package ensures thread safety through a read-write mutex, allowing
// concurrent reads while serializing writes.
package state
