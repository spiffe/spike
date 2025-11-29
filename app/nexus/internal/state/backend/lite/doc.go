//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lite provides an encryption-only backend implementation for SPIKE
// Nexus.
//
// This backend provides AES-GCM encryption and decryption services without
// persisting any data. It embeds the noop backend for storage operations
// (which are no-ops) and only implements the cryptographic interface.
//
// Use this backend when:
//   - Secrets are stored externally (e.g., in S3-compatible storage)
//   - SPIKE Nexus only needs to provide encrypt/decrypt operations
//   - No local secret or policy persistence is required
//
// In this mode, SPIKE policies are minimally enforced since the backend
// has no knowledge of the stored secrets.
//
// For full secret management with persistence, use the sqlite or memory
// backends instead.
package lite
