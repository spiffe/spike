//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package recovery implements root key management and disaster recovery for
// SPIKE Nexus using Shamir's Secret Sharing scheme.
//
// # Overview
//
// SPIKE Nexus encrypts all secrets with a root key. This package handles:
//   - Splitting the root key into shards distributed to SPIKE Keeper instances
//   - Reconstructing the root key from shards during startup or recovery
//   - Generating recovery shards for operators (manual disaster recovery)
//
// # Recovery Mechanisms
//
// The package supports two recovery paths:
//
// Automatic Recovery (via Keepers):
// When SPIKE Nexus starts, it contacts SPIKE Keeper instances to collect
// shards and reconstruct the root key. This happens automatically without
// operator intervention. Use InitializeBackingStoreFromKeepers for this path.
//
// Manual Recovery (via Operators):
// If automatic recovery fails (e.g., Keepers are unavailable), operators can
// manually restore the root key by providing recovery shards obtained earlier
// via NewPilotRecoveryShards. Use RestoreBackingStoreFromPilotShards for this
// path.
//
// # Key Functions
//
//   - InitializeBackingStoreFromKeepers: Recovers root key from Keeper shards
//   - RestoreBackingStoreFromPilotShards: Recovers root key from operator shards
//   - SendShardsPeriodically: Distributes shards to Keepers on a schedule
//   - NewPilotRecoveryShards: Generates recovery shards for operators
//   - ComputeRootKeyFromShards: Reconstructs root key using Shamir's algorithm
//
// # Security Considerations
//
// All shard data is zeroed from memory after use. Functions that encounter
// unrecoverable errors call log.FatalErr to prevent operation with corrupted
// cryptographic material.
package recovery
