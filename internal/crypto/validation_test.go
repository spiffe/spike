//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
)

// createValidShares creates a valid set of Shamir shares for testing.
// It uses the P256 group with a random secret.
func createValidShares(t *testing.T, threshold, numShares uint) (
	group.Scalar, []shamir.Share,
) {
	t.Helper()

	g := group.P256
	secret := g.RandomScalar(rand.Reader)

	ss := shamir.New(rand.Reader, threshold, secret)
	shares := ss.Share(numShares)

	return secret, shares
}

func TestVerifyShamirReconstruction_ValidShares(t *testing.T) {
	// Set environment to use panic instead of os.Exit
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set Shamir threshold to match our test
	t.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", "2")

	secret, shares := createValidShares(t, 1, 3) // threshold-1=1, so t=2

	// Should not panic for valid shares
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("VerifyShamirReconstruction() panicked unexpectedly: %v", r)
		}
	}()

	VerifyShamirReconstruction(secret, shares)
}

func TestVerifyShamirReconstruction_InvalidShares_RecoveryFails(t *testing.T) {
	// Set environment to use panic instead of os.Exit
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set Shamir threshold
	t.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", "2")

	// Create valid shares first
	_, shares := createValidShares(t, 1, 3)

	// Corrupt one of the shares to make recovery fail
	g := group.P256
	corruptedShares := make([]shamir.Share, len(shares))
	copy(corruptedShares, shares)
	corruptedShares[0].Value = g.RandomScalar(rand.Reader) // Corrupt the value

	// Use a different secret that won't match
	differentSecret := g.RandomScalar(rand.Reader)

	defer func() {
		if r := recover(); r == nil {
			t.Error("VerifyShamirReconstruction() should have panicked " +
				"for mismatched secret")
		}
	}()

	VerifyShamirReconstruction(differentSecret, corruptedShares)

	t.Error("Should have panicked before reaching here")
}

func TestVerifyShamirReconstruction_WrongSecret(t *testing.T) {
	// Set environment to use panic instead of os.Exit
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set Shamir threshold
	t.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", "2")

	// Create valid shares with one secret
	_, shares := createValidShares(t, 1, 3)

	// Try to verify with a different secret
	g := group.P256
	wrongSecret := g.RandomScalar(rand.Reader)

	defer func() {
		if r := recover(); r == nil {
			t.Error("VerifyShamirReconstruction() should have panicked " +
				"for wrong secret")
		}
	}()

	VerifyShamirReconstruction(wrongSecret, shares)

	t.Error("Should have panicked before reaching here")
}

func TestVerifyShamirReconstruction_InsufficientShares(t *testing.T) {
	// Set environment to use panic instead of os.Exit
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set threshold higher than available shares
	t.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", "5")

	// Create only 3 shares but require 5
	secret, shares := createValidShares(t, 4, 3) // Need 5 shares, only have 3

	defer func() {
		if r := recover(); r == nil {
			t.Error("VerifyShamirReconstruction() should have panicked " +
				"for insufficient shares")
		}
	}()

	VerifyShamirReconstruction(secret, shares)

	t.Error("Should have panicked before reaching here")
}
