//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package operator provides CLI commands for SPIKE Nexus administrative
// operations.
//
// This package implements privileged commands that require special SPIFFE roles
// for disaster recovery scenarios. These commands are used by operators to
// manage the root key recovery process.
//
// Available commands:
//
//   - recover: Retrieves recovery shards from a healthy SPIKE Nexus instance.
//     Requires the "recover" role. Shards are saved to the recovery directory
//     as hex-encoded files.
//
//   - restore: Submits recovery shards to a failed SPIKE Nexus instance to
//     restore the root key. Requires the "restore" role. Shards are entered
//     interactively with hidden input for security.
//
// Recovery workflow:
//
//  1. While SPIKE Nexus is healthy, run "spike operator recover" to obtain
//     and securely store the recovery shards.
//  2. If SPIKE Nexus fails and cannot auto-recover via SPIKE Keeper, run
//     "spike operator restore" multiple times to submit the required number
//     of shards.
//  3. Once enough shards are collected, SPIKE Nexus reconstructs the root key
//     and resumes normal operation.
//
// Security considerations:
//
//   - Recovery shards are sensitive cryptographic material and should be
//     encrypted and stored in separate secure locations.
//   - Input during restore is hidden to prevent shoulder-surfing.
//   - Shards are cleared from memory immediately after use.
//   - Role-based access control ensures only authorized operators can perform
//     these operations.
//
// See https://spike.ist/operations/recovery/ for detailed recovery procedures.
package operator
