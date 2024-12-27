//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package crypto

import "crypto/sha256"

type DeterministicReader struct {
	data []byte
	pos  int
}

func (r *DeterministicReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		// Generate more deterministic data if needed
		hash := sha256.Sum256(r.data)
		r.data = hash[:]
		r.pos = 0
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func NewDeterministicReader(seed []byte) *DeterministicReader {
	hash := sha256.Sum256(seed)
	return &DeterministicReader{
		data: hash[:],
		pos:  0,
	}
}
