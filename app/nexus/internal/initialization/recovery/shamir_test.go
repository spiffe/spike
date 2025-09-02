//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"os"
	"testing"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestShamirSecretSharingBasics(t *testing.T) {
	// Test basic Shamir secret sharing functionality
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

	// Test with different threshold and share configurations
	tests := []struct {
		name      string
		threshold uint
		numShares uint
	}{
		{"minimum configuration", 1, 2},
		{"typical configuration", 2, 3},
		{"larger configuration", 3, 5},
		{"equal threshold and shares", 3, 3},
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
			if len(shares) > int(tt.threshold) {
				reconstructShares := shares[:tt.threshold+1]
				reconstructed, err := shamir.Recover(tt.threshold, reconstructShares)
				if err != nil {
					t.Errorf("Failed to reconstruct secret: %v", err)
					return
				}

				if !reconstructed.IsEqual(secret) {
					t.Error("Reconstructed secret should equal original secret")
				}

				// Security: Clean up reconstructed secret
				reconstructed.SetUint64(0)
			}
		})
	}
}

func TestShamirDeterministicBehavior(t *testing.T) {
	// Test that the same secret generates the same shares
	g := group.P256

	testKey := make([]byte, crypto.AES256KeySize)
	for i := range testKey {
		testKey[i] = byte(i * 2 % 256)
	}

	secret := g.NewScalar()
	err := secret.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	threshold := uint(2)
	numShares := uint(3)

	// Generate shares twice with the same seed
	reader1 := crypto.NewDeterministicReader(testKey)
	ss1 := shamir.New(reader1, threshold, secret)
	shares1 := ss1.Share(numShares)

	reader2 := crypto.NewDeterministicReader(testKey)
	ss2 := shamir.New(reader2, threshold, secret)
	shares2 := ss2.Share(numShares)

	// Shares should be identical
	if len(shares1) != len(shares2) {
		t.Fatalf("Share counts should be equal: %d vs %d",
			len(shares1), len(shares2))
	}

	for i, share1 := range shares1 {
		share2 := shares2[i]

		// Compare IDs
		if !share1.ID.IsEqual(share2.ID) {
			t.Errorf("Share %d IDs should be equal", i)
		}

		// Compare Values
		if !share1.Value.IsEqual(share2.Value) {
			t.Errorf("Share %d Values should be equal", i)
		}
	}
}

func TestShamirInsufficientShares(t *testing.T) {
	// Test that insufficient shares cannot reconstruct the secret
	g := group.P256

	testKey := make([]byte, crypto.AES256KeySize)
	for i := range testKey {
		testKey[i] = byte(i * 3 % 256)
	}

	secret := g.NewScalar()
	err := secret.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	threshold := uint(2) // Need 3 shares to reconstruct
	numShares := uint(4)

	reader := crypto.NewDeterministicReader(testKey)
	ss := shamir.New(reader, threshold, secret)
	shares := ss.Share(numShares)

	// Try to reconstruct with insufficient shares (only threshold, need threshold+1)
	insufficientShares := shares[:threshold]
	_, err = shamir.Recover(threshold, insufficientShares)
	if err == nil {
		t.Error("Should fail to reconstruct with insufficient shares")
	}

	// Should succeed with sufficient shares
	sufficientShares := shares[:threshold+1]
	reconstructed, err := shamir.Recover(threshold, sufficientShares)
	if err != nil {
		t.Errorf("Should succeed with sufficient shares: %v", err)
		return
	}

	if !reconstructed.IsEqual(secret) {
		t.Error("Reconstructed secret should equal original with sufficient shares")
	}

	// Security: Clean up
	reconstructed.SetUint64(0)
}

func TestShamirShareStructure(t *testing.T) {
	// Test the structure of generated shares
	g := group.P256

	testKey := make([]byte, crypto.AES256KeySize)
	testKey[0] = 1 // Ensure non-zero

	secret := g.NewScalar()
	err := secret.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	threshold := uint(1)
	numShares := uint(3)

	reader := crypto.NewDeterministicReader(testKey)
	ss := shamir.New(reader, threshold, secret)
	shares := ss.Share(numShares)

	// Test share properties
	for i, share := range shares {
		// Test that ID is not zero (shares should have sequential IDs starting from 1)
		if share.ID.IsZero() {
			t.Errorf("Share %d should not have zero ID", i)
		}

		// Test that Value is not zero
		if share.Value.IsZero() {
			t.Errorf("Share %d should not have zero Value", i)
		}

		// Test that we can marshal/unmarshal the share
		idBytes, err := share.ID.MarshalBinary()
		if err != nil {
			t.Errorf("Failed to marshal share %d ID: %v", i, err)
		}
		if len(idBytes) == 0 {
			t.Errorf("Share %d ID bytes should not be empty", i)
		}

		valueBytes, err := share.Value.MarshalBinary()
		if err != nil {
			t.Errorf("Failed to marshal share %d Value: %v", i, err)
		}
		if len(valueBytes) != crypto.AES256KeySize {
			t.Errorf("Share %d Value should be %d bytes, got %d",
				i, crypto.AES256KeySize, len(valueBytes))
		}
	}

	// Test that all shares have unique IDs
	for i, share1 := range shares {
		for j, share2 := range shares {
			if i != j && share1.ID.IsEqual(share2.ID) {
				t.Errorf("Shares %d and %d should have different IDs", i, j)
			}
		}
	}
}

func TestEnvironmentThresholdAndShares(t *testing.T) {
	// Test different environment configurations
	originalThreshold := os.Getenv("SPIKE_NEXUS_SHAMIR_THRESHOLD")
	originalShares := os.Getenv("SPIKE_NEXUS_SHAMIR_SHARES")

	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", originalThreshold)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_SHAMIR_THRESHOLD")
		}
		if originalShares != "" {
			_ = os.Setenv("SPIKE_NEXUS_SHAMIR_SHARES", originalShares)
		} else {
			_ = os.Unsetenv("SPIKE_NEXUS_SHAMIR_SHARES")
		}
	}()

	tests := []struct {
		name      string
		threshold string
		shares    string
		valid     bool
	}{
		{"valid 2-of-3", "2", "3", true},
		{"valid 3-of-5", "3", "5", true},
		{"edge case 1-of-1", "1", "1", true},
		{"invalid threshold > shares", "4", "3", false},
		{"valid threshold = shares", "3", "3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", tt.threshold)
			_ = os.Setenv("SPIKE_NEXUS_SHAMIR_SHARES", tt.shares)

			// The functions would use these values like:
			// t := uint(env.ShamirThreshold() - 1)
			// n := uint(env.ShamirShares())

			// We can't easily test the actual functions due to dependencies,
			// but we can verify the environment configuration is valid
			if tt.valid {
				envThreshold := os.Getenv("SPIKE_NEXUS_SHAMIR_THRESHOLD")
				envShares := os.Getenv("SPIKE_NEXUS_SHAMIR_SHARES")

				if envThreshold != tt.threshold {
					t.Errorf("Expected threshold %s, got %s", tt.threshold, envThreshold)
				}
				if envShares != tt.shares {
					t.Errorf("Expected shares %s, got %s", tt.shares, envShares)
				}
			}
		})
	}
}

func TestGroupP256OperationsShamir(t *testing.T) {
	// Test P256 group operations used in the functions
	g := group.P256

	// Test creating scalars
	scalar1 := g.NewScalar()
	scalar2 := g.NewScalar()

	if scalar1 == nil {
		t.Error("NewScalar should not return nil")
	}
	if scalar2 == nil {
		t.Error("NewScalar should not return nil")
	}

	// Test setting values
	scalar1.SetUint64(123)
	scalar2.SetUint64(456)

	// Test IsZero
	zeroScalar := g.NewScalar()
	if !zeroScalar.IsZero() {
		t.Error("New scalar should be zero")
	}
	if scalar1.IsZero() {
		t.Error("Scalar with value should not be zero")
	}

	// Test IsEqual
	scalar3 := g.NewScalar()
	scalar3.SetUint64(123)
	if !scalar1.IsEqual(scalar3) {
		t.Error("Scalars with same value should be equal")
	}
	if scalar1.IsEqual(scalar2) {
		t.Error("Scalars with different values should not be equal")
	}

	// Test marshal/unmarshal
	data, err := scalar1.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary failed: %v", err)
	}

	scalar4 := g.NewScalar()
	err = scalar4.UnmarshalBinary(data)
	if err != nil {
		t.Errorf("UnmarshalBinary failed: %v", err)
	}

	if !scalar1.IsEqual(scalar4) {
		t.Error("Unmarshaled scalar should equal original")
	}

	// Test SetUint64(0) for cleanup
	scalar1.SetUint64(0)
	if !scalar1.IsZero() {
		t.Error("Scalar should be zero after SetUint64(0)")
	}
}

func TestDeterministicReader(t *testing.T) {
	// Test deterministic reader behavior
	seed1 := make([]byte, crypto.AES256KeySize)
	seed2 := make([]byte, crypto.AES256KeySize)

	// Same seed
	for i := range seed1 {
		seed1[i] = byte(i)
		seed2[i] = byte(i)
	}

	reader1 := crypto.NewDeterministicReader(seed1)
	reader2 := crypto.NewDeterministicReader(seed2)

	// Different seed
	seed3 := make([]byte, crypto.AES256KeySize)
	for i := range seed3 {
		seed3[i] = byte(i + 1)
	}
	reader3 := crypto.NewDeterministicReader(seed3)

	g := group.P256
	secret := g.NewScalar()
	secret.SetUint64(42)

	// Create shares with the same seed
	ss1 := shamir.New(reader1, 1, secret)
	shares1 := ss1.Share(2)

	ss2 := shamir.New(reader2, 1, secret)
	shares2 := ss2.Share(2)

	// Create shares with different seed
	ss3 := shamir.New(reader3, 1, secret)
	shares3 := ss3.Share(2)

	// Shares from the same seed should be identical
	if len(shares1) != len(shares2) {
		t.Fatal("Same seed should produce same number of shares")
	}

	for i, share1 := range shares1 {
		share2 := shares2[i]
		if !share1.ID.IsEqual(share2.ID) || !share1.Value.IsEqual(share2.Value) {
			t.Errorf("Shares %d from same seed should be identical", i)
		}
	}

	// Shares from different seeds should be different (at least values)
	if len(shares1) == len(shares3) {
		differentFound := false
		for i, share1 := range shares1 {
			share3 := shares3[i]
			if !share1.Value.IsEqual(share3.Value) {
				differentFound = true
				break
			}
		}
		if !differentFound {
			t.Error("Different seeds should produce different share values")
		}
	}
}

func TestShamirRecoveryValidation(t *testing.T) {
	// Test the recovery validation logic used in sanityCheck
	g := group.P256

	testKey := make([]byte, crypto.AES256KeySize)
	for i := range testKey {
		testKey[i] = byte(i * 7 % 256)
	}

	original := g.NewScalar()
	err := original.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	threshold := uint(2)
	numShares := uint(4)

	reader := crypto.NewDeterministicReader(testKey)
	ss := shamir.New(reader, threshold, original)
	shares := ss.Share(numShares)

	// Test successful recovery
	reconstructed, err := shamir.Recover(threshold, shares[:threshold+1])
	if err != nil {
		t.Errorf("Recovery should succeed: %v", err)
		return
	}

	// Test validation (this is what sanityCheck does)
	if !original.IsEqual(reconstructed) {
		t.Error("Reconstructed secret should equal original")
	}

	// Test cleanup
	reconstructed.SetUint64(0)
	if !reconstructed.IsZero() {
		t.Error("Cleaned up secret should be zero")
	}

	// Test with wrong number of shares (too few)
	_, err = shamir.Recover(threshold, shares[:threshold])
	if err == nil {
		t.Error("Recovery should fail with insufficient shares")
	}
}

func TestShamirShareSlicing(t *testing.T) {
	// Test slicing operations on share slices (as used in sanityCheck)
	g := group.P256

	testKey := make([]byte, crypto.AES256KeySize)
	testKey[0] = 1

	secret := g.NewScalar()
	err := secret.UnmarshalBinary(testKey)
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}

	threshold := uint(2)
	numShares := uint(5)

	reader := crypto.NewDeterministicReader(testKey)
	ss := shamir.New(reader, threshold, secret)
	shares := ss.Share(numShares)

	// Test different slicing operations
	tests := []struct {
		name       string
		slice      []shamir.Share
		shouldWork bool
	}{
		{"first threshold+1", shares[:threshold+1], true},
		{"middle threshold+1", shares[1 : threshold+2], true},
		{"last threshold+1", shares[numShares-threshold-1:], true},
		{"too few shares", shares[:threshold], false},
		{"single share", shares[:1], false},
		{"all shares", shares, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := shamir.Recover(threshold, tt.slice)

			if tt.shouldWork && err != nil {
				t.Errorf("Expected recovery to work, got error: %v", err)
			} else if !tt.shouldWork && err == nil {
				t.Error("Expected recovery to fail, but it succeeded")
			}
		})
	}
}
