//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package bootstrap provides HTTP route handlers for SPIKE Bootstrap
// verification endpoints.
//
// This package implements the proof-of-possession (PoP) verification endpoint
// that allows SPIKE Bootstrap to confirm that SPIKE Nexus has been properly
// initialized with the root key.
//
// # Verification Flow
//
// During the bootstrap sequence, SPIKE Bootstrap needs to verify that Nexus
// successfully received and initialized the root key. The verification endpoint
// implements a challenge-response protocol:
//
//  1. Bootstrap sends encrypted data (ciphertext encrypted with the root key)
//  2. Nexus decrypts the data using its root key
//  3. Nexus returns the decrypted plaintext
//  4. Bootstrap verifies the plaintext matches the original data
//
// This proves Nexus possesses the correct root key without exposing the key
// itself over the network.
//
// # Security Constraints
//
// The verification endpoint is restricted to SPIKE Bootstrap only. The guard
// function validates that the caller's SPIFFE ID matches the expected SPIKE
// Bootstrap identity before processing the verification request.
//
// # Authentication
//
// All endpoints require mTLS authentication via SPIFFE. The caller's SPIFFE ID
// is extracted from the client certificate and validated before any operations
// are performed.
package bootstrap
