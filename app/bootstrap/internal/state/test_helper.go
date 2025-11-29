//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"testing"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"
)

// resetRootSharesForTesting resets the rootSharesGenerated flag to allow
// multiple calls to RootShares() within tests. This function should ONLY be
// used in test code to enable testing of RootShares() behavior.
//
// WARNING: This function should never be called in production code.
func resetRootSharesForTesting() {
	rootSharesGeneratedMu.Lock()
	rootSharesGenerated = false
	rootSharesGeneratedMu.Unlock()
}

// Helper function to create test shares with known structure
func createTestShares(t *testing.T, numShares int) []shamir.Share {
	g := group.P256

	// Create a test secret
	secret := g.NewScalar()
	testKey := make([]byte, crypto.AES256KeySize)
	for i := range testKey {
		testKey[i] = byte(i % 256)
	}

	err := secret.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	// Create shares with threshold = numShares - 1
	threshold := uint(numShares - 1)
	reader := crypto.NewDeterministicReader(testKey)
	ss := shamir.New(reader, threshold, secret)

	return ss.Share(uint(numShares))
}
