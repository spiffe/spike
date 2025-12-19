//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package crypto

// GCMNonceSize is the standard nonce size for AES-GCM as recommended by
// NIST SP 800-38D. This is 12 bytes (96 bits).
//
// While GCM technically supports other nonce sizes via NewGCMWithNonceSize(),
// the 12-byte standard is strongly preferred because:
//   - It uses the more efficient counter mode internally
//   - Non-standard sizes require additional GHASH operations
//   - It is the NIST-recommended size for maximum interoperability
//   - Go's cipher.NewGCM() uses this size by default
//
// This constant is used for validation of incoming nonces in the cipher API
// and bootstrap verification endpoints. See ADR-0032 for the design decision.
// (https://spike.ist/architecture/adrs/adr-0032/)
const GCMNonceSize = 12
