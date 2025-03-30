//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
)

type ShamirShard struct {
	Id    uint64
	Value *[32]byte
}

// RecoverRootKey reconstructs the original root key from a slice of ShamirShard.
// It uses Shamir's Secret Sharing scheme to recover the original secret.
//
// Parameters:
//   - ss []ShamirShard: A slice of ShamirShard structures, each containing
//     an ID and a pointer to a 32-byte value representing a secret share
//
// Returns:
//   - *[32]byte: A pointer to the reconstructed 32-byte root key
//
// The function will:
//   - Convert each ShamirShard into a properly formatted secretsharing.Share
//   - Use the IDs from the provided ShamirShards
//   - Retrieve the threshold from the environment
//   - Reconstruct the original secret using the secretsharing.Recover function
//   - Validate the recovered key has the correct length (32 bytes)
//   - Zero out all shares after use for security
//
// It will log a fatal error and exit if:
//   - Any share fails to unmarshal properly
//   - The recovery process fails
//   - The reconstructed key is nil
//   - The binary representation has an incorrect length
func RecoverRootKey(ss []ShamirShard) *[32]byte {
	const fName = "RecoverRootKey"

	g := group.P256
	shares := make([]secretsharing.Share, 0, len(ss))
	// Security: Ensure that the shares are zeroed out after the function returns:
	defer func() {
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}
	}()

	// Process all provided shares
	for _, shamirShard := range ss {
		// Create a new share with sequential Id (starting from 1)
		share := secretsharing.Share{
			ID:    g.NewScalar(),
			Value: g.NewScalar(),
		}

		// Set ID
		share.ID.SetUint64(shamirShard.Id)

		// Unmarshal the binary data
		err := share.Value.UnmarshalBinary(shamirShard.Value[:])
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
		// Security: reset shares.
		// defer won't get called since log.Fatalln terminates the program.
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}

		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}

	if reconstructed == nil {
		// Security: reset shares.
		// defer won't get called since log.Fatalln terminates the program.
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}

		log.FatalLn(fName + ": Failed to reconstruct the root key")
	}

	if reconstructed != nil {
		binaryRec, err := reconstructed.MarshalBinary()
		if err != nil {
			// Security: Zero out:
			reconstructed.SetUint64(0)

			log.FatalLn(fName + ": Failed to marshal: " + err.Error())
			return &[32]byte{}
		}

		if len(binaryRec) != 32 {
			log.FatalLn(fName + ": Reconstructed root key has incorrect length")
			return &[32]byte{}
		}

		var result [32]byte
		copy(result[:], binaryRec)
		// Security: Zero out temporary variables before function exits.
		mem.Clear(&binaryRec)

		return &result
	}

	return &[32]byte{}
}
