//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

const spikeCipherVersion = byte('1')
const headerKeyContentType = "Content-Type"
const headerValueOctetStream = "application/octet-stream"

// AES-GCM standard nonce size is 12 bytes
const expectedNonceSize = 12

// Maximum ciphertext size to prevent DoS attacks.
// The theoretical maximum for GCM is 68,719,476,704 bytes (limited by the
// 32-bit counter), but we set a much lower practical limit.
const maxCiphertextSize = 65536

// Maximum plaintext size to prevent DoS attacks.
// This is set to account for the AES-GCM authentication tag overhead
// (16 bytes), so the resulting ciphertext stays within maxCiphertextSize.
const maxPlaintextSize = maxCiphertextSize - 16
