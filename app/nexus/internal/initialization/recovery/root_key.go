//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

type ShamirShard struct {
	ID    uint64
	Value *[crypto.AES256KeySize]byte
}

// ComputeRootKeyFromShards reconstructs the original root key from a slice of
// ShamirShard. It uses Shamir's Secret Sharing scheme to recover the original
// secret.
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
func ComputeRootKeyFromShards(ss []ShamirShard) *[crypto.AES256KeySize]byte {
	const fName = "ComputeRootKeyFromShards"

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
		// Create a new share with sequential ID (starting from 1):
		share := secretsharing.Share{
			ID:    g.NewScalar(),
			Value: g.NewScalar(),
		}

		// Set ID
		share.ID.SetUint64(shamirShard.ID)

		// Unmarshal the binary data
		unmarshalErr := share.Value.UnmarshalBinary(shamirShard.Value[:])
		if unmarshalErr != nil {
			failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(unmarshalErr)
			failErr.Msg = "failed to unmarshal shard"
			log.FatalErr(fName, *failErr)
		}

		shares = append(shares, share)
	}

	// Recover the secret
	// The first parameter to Recover is threshold-1
	// We need the threshold from the environment
	threshold := env.ShamirThresholdVal()
	reconstructed, recoverErr := secretsharing.Recover(uint(threshold-1), shares)
	if recoverErr != nil {
		// Security: Reset shares.
		// Defer won't get called because log.FatalErr terminates the program.
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}

		failErr := sdkErrors.ErrShamirReconstructionFailed.Wrap(recoverErr)
		failErr.Msg = "failed to recover secret"
		log.FatalErr(fName, *failErr)
	}

	if reconstructed == nil {
		// Security: Reset shares.
		// Defer won't get called because log.FatalErr terminates the program.
		for _, s := range shares {
			s.ID.SetUint64(0)
			s.Value.SetUint64(0)
		}

		failErr := *sdkErrors.ErrShamirReconstructionFailed.Clone()
		failErr.Msg = "failed to reconstruct the root key"
		log.FatalErr(fName, failErr)
	}

	if reconstructed != nil {
		binaryRec, marshalErr := reconstructed.MarshalBinary()
		if marshalErr != nil {
			// Security: Zero out:
			reconstructed.SetUint64(0)

			failErr := sdkErrors.ErrDataMarshalFailure.Wrap(marshalErr)
			failErr.Msg = "failed to marshal reconstructed key"
			log.FatalErr(fName, *failErr)

			return &[crypto.AES256KeySize]byte{}
		}

		if len(binaryRec) != crypto.AES256KeySize {
			failErr := *sdkErrors.ErrDataInvalidInput.Clone()
			failErr.Msg = "reconstructed root key has incorrect length"
			log.FatalErr(fName, failErr)

			return &[crypto.AES256KeySize]byte{}
		}

		var result [crypto.AES256KeySize]byte
		copy(result[:], binaryRec)
		// Security: Zero out temporary variables before the function exits.
		mem.ClearBytes(binaryRec)

		return &result
	}

	return &[crypto.AES256KeySize]byte{}
}
