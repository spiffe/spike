//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package kek provides Key Encryption Key (KEK) management functionality
// for SPIKE Nexus. It implements envelope encryption where:
//   - Each secret has a random Data Encryption Key (DEK)
//   - DEKs are wrapped by versioned KEKs
//   - KEKs are deterministically derived from the Root Master Key (RMK) using HKDF
//
// # Architecture
//
// The key hierarchy is:
//
//	Secret Data → DEK (random per secret) → KEK (versioned) → RMK (root)
//
// # Key Features
//
// KEK Rotation:
//   - Time-based: Rotate after N days (default: 90)
//   - Usage-based: Rotate after N wraps (default: 20M)
//   - Grace period: Old KEKs remain readable for N days (default: 180)
//   - Automatic scheduler checks rotation conditions periodically
//
// Lazy Rewrapping:
//   - Secrets are rewrapped on-demand when accessed
//   - Minimizes downtime during rotation
//   - Background sweeper handles cold data
//   - Configurable QPS limits prevent overload
//
// RMK Rotation:
//   - Only rewraps KEKs, not secrets
//   - Preserves KEK values through deterministic derivation
//   - Snapshot and validation support
//   - No bulk data re-encryption needed
//
// # Configuration
//
// Environment variables (all optional):
//   - SPIKE_KEK_ROTATION_ENABLED: Enable KEK rotation (default: false)
//   - SPIKE_KEK_ROTATION_DAYS: Days before rotation (default: 90)
//   - SPIKE_KEK_MAX_WRAPS: Max wraps before rotation (default: 20000000)
//   - SPIKE_KEK_GRACE_DAYS: Grace period days (default: 180)
//   - SPIKE_KEK_LAZY_REWRAP_ENABLED: Enable lazy rewrap (default: true)
//   - SPIKE_KEK_MAX_REWRAP_QPS: Max rewrap rate (default: 100)
//
// # API Endpoints
//
//   - POST /v1/kek/rotate - Manually trigger KEK rotation
//   - GET /v1/kek/current - Get current active KEK info
//   - GET /v1/kek/list - List all KEKs with status
//   - GET /v1/kek/stats - Get rotation statistics
//   - POST /v1/rmk/rotate - Initiate RMK rotation (manual ceremony)
//   - GET /v1/rmk/snapshot - Create RMK rotation snapshot
//
// # Security Properties
//
//   - KEKs never stored in plaintext (derived on-demand from RMK)
//   - DEKs only exist in memory during encryption/decryption
//   - Domain separation prevents KEK reuse across contexts
//   - Cryptographic agility through AEAD algorithm field
//
// # Example Usage
//
//	// Initialize KEK manager
//	manager, err := kek.NewManager(rmk, rmkVersion, policy, storage)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Wrap a DEK
//	dek, _ := kek.GenerateDEK()
//	wrappedDEK, _ := kek.WrapDEK(dek, kek, kekID)
//
//	// Unwrap a DEK
//	unwrappedDEK, _ := kek.UnwrapDEK(wrappedDEK, kek)
//
//	// Check if rotation needed
//	if manager.ShouldRotate() {
//		manager.RotateKEK()
//	}
package kek
