//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
)

// RecoverRootKey reconstructs the original root key from a slice of secret
// shares. It uses Shamir's Secret Sharing scheme to recover the original
// secret.
//
// Parameters:
//   - ss: A slice of byte slices, where each byte slice represents a secret
//     share
//
// Returns:
//   - []byte: The reconstructed root key as a byte slice
//
// The function will:
//   - Convert each raw byte slice into a properly formatted secretsharing.Share
//   - Assign sequential IDs to each share starting from 1
//   - Reconstruct the original secret using the secretsharing.Recover function
//   - Validate the recovered key has the correct length (32 bytes)
//
// It will log a fatal error and exit if:
//   - Any share fails to unmarshal properly
//   - The recovery process fails
//   - The reconstructed key is nil
//   - The binary representation has an incorrect length
func RecoverRootKey(ss []*[32]byte) *[32]byte {
	const fName = "RecoverRootKey"

	g := group.P256
	shares := make([]secretsharing.Share, 0, len(ss))
	defer func() {
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}
	}()

	// Process all provided shares
	for i, shareBinary := range ss {
		// Create a new share with sequential ID (starting from 1)
		share := secretsharing.Share{
			ID:    g.NewScalar(),
			Value: g.NewScalar(),
		}

		// Set ID (1-indexed)
		share.ID.SetUint64(uint64(i + 1))

		// Unmarshal the binary data
		err := share.Value.UnmarshalBinary(shareBinary[:])
		if err != nil {
			log.FatalLn(fName + ": Failed to unmarshal share: " + err.Error())
		}

		shares = append(shares, share)
	}

	// Recover the secret
	// The first parameter to Recover is threshold-1
	// We need the threshold from the environment
	threshold := env.ShamirThreshold()
	reconstructed, err := secretsharing.Recover(uint(threshold-1), shares)
	if err != nil {
		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}

	if reconstructed == nil {
		log.FatalLn(fName + ": Failed to reconstruct the root key")
	}

	if reconstructed != nil {
		binaryRec, err := reconstructed.MarshalBinary()
		if err != nil {
			log.FatalLn(fName + ": Failed to marshal: " + err.Error())
			return &[32]byte{}
		}

		if len(binaryRec) != 32 {
			log.FatalLn(fName + ": Reconstructed root key has incorrect length")
			return &[32]byte{}
		}

		var result [32]byte
		for i := range binaryRec {
			result[i] = binaryRec[i]
		}
		for i := range binaryRec {
			binaryRec[i] = 0
		}

		return &result
	}

	return &[32]byte{}
}
