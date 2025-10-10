//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package cipher provides HTTP route handlers for encryption and decryption
// operations in SPIKE Nexus.
//
// This package implements encryption-as-a-service endpoints that allow
// workloads to encrypt and decrypt data using the Nexus cipher without
// persisting the data. It supports both JSON and streaming modes for
// efficient handling of different data types and sizes.
package cipher
