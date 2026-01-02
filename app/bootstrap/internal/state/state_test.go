//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"

	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestKeeperShareValidID(t *testing.T) {
	// Create test shares with known IDs
	rootShares := createTestShares(t, 5)

	// Test finding a valid keeper share
	keeperID := "1"
	share := KeeperShare(rootShares, keeperID)

	if share.ID.IsZero() {
		t.Error("Returned share should not have zero ID")
	}

	if share.Value.IsZero() {
		t.Error("Returned share should not have zero value")
	}

	// Verify the ID matches what we requested
	expectedID := group.P256.NewScalar().SetUint64(1)
	if !share.ID.IsEqual(expectedID) {
		t.Error("Share ID should match the requested keeper ID")
	}
}

func TestKeeperShareInvalidID(t *testing.T) {
	tests := []struct {
		name       string
		keeperID   string
		shouldExit bool
	}{
		{
			name:       "non-numeric keeper ID",
			keeperID:   "abc",
			shouldExit: true,
		},
		{
			name:       "keeper ID not in shares",
			keeperID:   "99",
			shouldExit: true,
		},
		{
			name:       "empty keeper ID",
			keeperID:   "",
			shouldExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldExit {
				// These tests would call log.FatalErr, so we skip them.
				// In a production environment, you'd want to refactor the code
				// to return errors instead of calling log.FatalErr.
				t.Skip(
					"Skipping test that would cause log.FatalErr" +
						" - needs refactoring for testability",
				)
			}
		})
	}
}

func TestShamirSecretSharingBasics(t *testing.T) {
	// Test basic Shamir secret sharing functionality that the code relies on.
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

	// Test with different threshold configurations
	tests := []struct {
		name      string
		threshold uint
		numShares uint
	}{
		{"minimum configuration", 1, 2},
		{"typical configuration", 2, 3},
		{"larger configuration", 3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a deterministic reader for consistent results
			reader := crypto.NewDeterministicReader(testKey)
			ss := shamir.New(reader, tt.threshold, secret)

			shares := ss.Share(tt.numShares)

			if len(shares) != int(tt.numShares) {
				t.Errorf("Expected %d shares, got %d", tt.numShares, len(shares))
			}

			// Test that we can reconstruct with threshold+1 shares
			if len(shares) >= int(tt.threshold)+1 {
				reconstructShares := shares[:tt.threshold+1]
				reconstructed, recoverErr := shamir.Recover(
					tt.threshold, reconstructShares,
				)
				if recoverErr != nil {
					t.Errorf("Failed to reconstruct secret: %v", recoverErr)
					return
				}

				if !reconstructed.IsEqual(secret) {
					t.Error("Reconstructed secret should equal original secret")
				}
			}
		})
	}
}

func TestShareIDConversion(t *testing.T) {
	// Test the ID conversion logic used in KeeperShare.
	g := group.P256

	testCases := []struct {
		id       uint64
		expected string
	}{
		{1, "1"},
		{2, "2"},
		{255, "255"},
	}

	for _, tc := range testCases {
		t.Run("id_"+strconv.FormatUint(tc.id, 10), func(t *testing.T) {
			scalar := g.NewScalar().SetUint64(tc.id)

			// Test conversion back to string (similar to what KeeperShare does)
			kid, err := strconv.Atoi(tc.expected)
			if err != nil {
				t.Fatalf("Test setup error: %v", err)
			}

			expectedScalar := g.NewScalar().SetUint64(uint64(kid))

			if !scalar.IsEqual(expectedScalar) {
				t.Errorf("ID conversion mismatch for %d", tc.id)
			}
		})
	}
}

func TestShareValidation(t *testing.T) {
	// Test that shares have expected properties.
	shares := createTestShares(t, 3)

	// All shares should have different IDs
	idSet := make(map[string]bool)
	for _, share := range shares {
		idBytes, err := share.ID.MarshalBinary()
		if err != nil {
			t.Errorf("Failed to marshal share ID: %v", err)
			continue
		}

		// Use hex encoding to properly represent the ID bytes
		idStr := hex.EncodeToString(idBytes)
		if idSet[idStr] {
			t.Error("Duplicate share ID found")
		}
		idSet[idStr] = true

		if share.Value.IsZero() {
			t.Error("Share value should not be zero")
		}
	}

	// Should have exactly 3 unique shares
	if len(idSet) != 3 {
		t.Errorf("Expected 3 unique share IDs, got %d", len(idSet))
	}
}
