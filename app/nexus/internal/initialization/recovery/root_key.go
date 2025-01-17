//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/hex"
	"sync"

	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"

	"github.com/spiffe/spike/internal/log"
)

var rootKey []byte
var rootKeyMu sync.RWMutex

func mustUpdateRecoveryInfo(rk string) []secretsharing.Share {
	decodedRootKey, err := hex.DecodeString(rk)
	if err != nil {
		log.FatalLn("Bootstrap: failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	// Save recovery information.
	rootKeyMu.Lock()
	rootKey = decodedRootKey
	rootKeyMu.Unlock()

	// TODO: we don't need these methods anymore
	//persist.AsyncPersistRecoveryInfo(store.KeyRecoveryData{

	return rootShares
}

func recoverRootKey(ss [][]byte) []byte {
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
		log.FatalLn("Failed to unmarshal share: " + err.Error())
	}
	secondShare := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	secondShare.ID.SetUint64(2)
	err = secondShare.Value.UnmarshalBinary(secondShard)
	if err != nil {
		log.FatalLn("Failed to unmarshal share: " + err.Error())
	}

	var shares []secretsharing.Share
	shares = append(shares, firstShare)
	shares = append(shares, secondShare)

	reconstructed, err := secretsharing.Recover(1, shares)
	if err != nil {
		log.FatalLn("Failed to recover: " + err.Error())
	}

	// TODO: this has assumption that there are 2 shards. it should not be.

	// TODO: check for errors.

	if reconstructed == nil {
		log.FatalLn("Failed to reconstruct the root key")
		return []byte{}
	}

	binaryRec, err := reconstructed.MarshalBinary()
	if err != nil {
		log.FatalLn("Failed to marshal: " + err.Error())
		return []byte{}
	}

	// TODO: check size 32bytes.

	return binaryRec
}
