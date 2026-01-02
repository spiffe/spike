//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import "github.com/spiffe/spike-sdk-go/crypto"

const spikeCipherVersion = byte('1')
const headerKeyContentType = "Content-Type"
const headerValueOctetStream = "application/octet-stream"

// expectedNonceSize is the standard AES-GCM nonce size. See ADR-0032.
// (https://spike.ist/architecture/adrs/adr-0012/)
const expectedNonceSize = crypto.GCMNonceSize
