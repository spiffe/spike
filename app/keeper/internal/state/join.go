//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

//func setInternalShard(shares []secretsharing.Share) {
//	// Sort the keys of env.Peers() alphabetically for deterministic
//	// shard indexing.
//	peers := env.Peers()
//	peerKeys := make([]string, 0, len(peers))
//	for id := range peers {
//		peerKeys = append(peerKeys, id)
//	}
//	sort.Strings(peerKeys)
//
//	myId := env.KeeperId()
//
//	// Find the index of the current Keeper's ID
//	var myShard []byte
//	for index, id := range peerKeys {
//		if id == myId {
//			// Save the shard corresponding to this Keeper
//			if val, ok := Shards.Load(myId); ok {
//				myShard = val.([]byte)
//
//				log.Log().Info("setInternalShard", "id", myId, "index", index)
//
//				shareVal, _ := shares[index].Value.MarshalBinary()
//
//				SetShard(shareVal)
//				EraseIntermediateShards()
//
//				break
//			}
//		}
//	}
//
//	// Ensure myShard is stored correctly in the state namespace
//	if myShard == nil {
//		log.FatalLn(
//			"setInternalShard: Shard for Keeper ID", myId, "could not be found",
//		)
//	}
//}

//func computeFinalKey() []byte {
//	finalKey := make([]byte, 32)
//
//	counter := 0
//	Shards.Range(func(key, value any) bool {
//		counter++
//		shard := value.([]byte)
//		for i := 0; i < 32; i++ {
//			finalKey[i] ^= shard[i]
//		}
//		return true
//	})
//
//	if counter != 3 {
//		log.FatalLn("computeFinalKey: Not all shards received")
//	}
//
//	if len(finalKey) != 32 {
//		log.FatalLn("computeFinalKey: FinalKey must be 32 bytes long")
//	}
//
//	return finalKey
//}

//func sanityCheck(secret group.Scalar, shares []secretsharing.Share) {
//	t := uint(1) // Need t+1 shares to reconstruct
//
//	reconstructed, err := secretsharing.Recover(t, shares[:2])
//	if err != nil {
//		log.FatalLn("computeShares: Failed to recover: " + err.Error())
//	}
//	if !secret.IsEqual(reconstructed) {
//		log.FatalLn("computeShares: Recovered secret does not match original")
//	}
//}
