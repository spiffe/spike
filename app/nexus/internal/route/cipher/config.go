//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

const spikeCipherVersion = byte('1')
const headerKeyContentType = "Content-Type"
const headerValueOctetStream = "application/octet-stream"

// AES-GCM standard nonce size is 12 bytes
const expectedNonceSize = 12
