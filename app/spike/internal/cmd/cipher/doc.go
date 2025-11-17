//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package cipher provides cryptographic operations for encrypting and
// decrypting data through SPIKE Nexus. It enables workloads to protect
// sensitive data in transit or at rest using SPIFFE-based authentication.
//
// The package implements CLI commands that support two operational modes:
//
// Stream Mode:
//   - Encrypts or decrypts data from files or stdin/stdout
//   - Handles binary data transparently
//   - Ideal for processing large files or piping data between commands
//
// JSON Mode:
//   - Works with base64-encoded data components
//   - Allows explicit control over cryptographic parameters (version, nonce)
//   - Useful for integration with external systems or APIs
//
// All operations authenticate with SPIKE Nexus using SPIFFE identities,
// ensuring that only authorized workloads can perform cryptographic
// operations. The actual encryption keys and algorithms are managed
// server-side by SPIKE Nexus.
//
// Key Features:
//   - SPIFFE-based authentication for all operations
//   - Support for file-based and stream-based encryption/decryption
//   - Flexible input/output modes (files, stdin/stdout, base64 strings)
//   - Server-side key management through SPIKE Nexus
//
// See https://spike.ist/usage/commands/ for CLI documentation.
package cipher
