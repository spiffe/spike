//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/hex"

	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
)

func mustUpdateRecoveryInfo(rk string) []secretsharing.Share {
	const fName = "mustUpdateRecoveryInfo"

	decodedRootKey, err := hex.DecodeString(rk)
	if err != nil {
		log.FatalLn(fName + ": failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	// Save recovery information.
	state.SetRootKey(decodedRootKey)

	return rootShares
}

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
