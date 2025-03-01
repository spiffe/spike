//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"

	"github.com/spiffe/spike/internal/log"
)

// RecoverRootKey reconstructs the original root key from a minimum of two
// secret shares.
//
// The function uses Shamir's Secret Sharing implemented in the secretsharing
// package to reconstruct the original secret (root key) from at least two
// shares. It works with the P256 elliptic curve as the mathematical group for
// the secret sharing operations.
//
// Parameters:
//   - ss [][]byte: A slice containing at least two byte slices representing
//     the secret shares.
//     The function currently uses only the first two shares (ss[0] and ss[1]).
//
// Returns:
//   - []byte: The reconstructed root key as a 32-byte slice. If reconstruction
//     fails, returns an empty byte slice after logging a fatal error.
//
// The function will log a fatal error and terminate execution if:
//   - It fails to unmarshal any of the shares
//   - The secret reconstruction operation fails
//   - The reconstructed secret is nil
//   - The marshaled reconstructed secret doesn't have the expected 32-byte
//     length
func RecoverRootKey(ss [][]byte) []byte {
	const fName = "RecoverRootKey"

	g := group.P256
	firstShard := ss[0]
	secondShard := ss[1]
	firstShare := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	firstShare.ID.SetUint64(1)
	err := firstShare.Value.UnmarshalBinary(firstShard)
	if err != nil {
		log.FatalLn(fName + ": Failed to unmarshal share: " + err.Error())
	}
	secondShare := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	secondShare.ID.SetUint64(2)
	err = secondShare.Value.UnmarshalBinary(secondShard)
	if err != nil {
		log.FatalLn(fName + ": Failed to unmarshal share: " + err.Error())
	}

	var shares []secretsharing.Share
	shares = append(shares, firstShare)
	shares = append(shares, secondShare)

	reconstructed, err := secretsharing.Recover(1, shares)
	if err != nil {
		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}

	if reconstructed == nil {
		log.FatalLn(fName + ": Failed to reconstruct the root key")
		return []byte{}
	}

	binaryRec, err := reconstructed.MarshalBinary()
	if err != nil {
		log.FatalLn(fName + ": Failed to marshal: " + err.Error())
		return []byte{}
	}

	if len(binaryRec) != 32 {
		log.FatalLn(fName + ": Reconstructed root key has incorrect length")
		return []byte{}
	}

	return binaryRec
}
