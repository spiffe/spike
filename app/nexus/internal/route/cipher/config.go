//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import "github.com/spiffe/spike/internal/crypto"

const spikeCipherVersion = byte('1')
const headerKeyContentType = "Content-Type"
const headerValueOctetStream = "application/octet-stream"

// expectedNonceSize is the standard AES-GCM nonce size. See ADR-0032.
const expectedNonceSize = crypto.GCMNonceSize
