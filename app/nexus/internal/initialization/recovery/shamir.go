//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"sort"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
)

func sanityCheck(secret group.Scalar, shares []shamir.Share) {
	const fName = "sanityCheck"

	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct

	reconstructed, err := shamir.Recover(t, shares[:env.ShamirThreshold()])
	if err != nil {
		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		log.FatalLn(fName + ": Recovered secret does not match original")
	}
}

func computeShares(finalKey []byte) (group.Scalar, []shamir.Share) {
	const fName = "computeShares"

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirShares())        // Total number of shares

	// Create secret from your 32 byte key
	secret := g.NewScalar()
	if err := secret.UnmarshalBinary(finalKey); err != nil {
		log.FatalLn(fName + ": Failed to unmarshal key: %v" + err.Error())
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(finalKey)
	ss := shamir.New(reader, t, secret)
	return secret, ss.Share(n)
}

func findShare(id string, keepers map[string]string,
	shares []shamir.Share,
) *shamir.Share {
	// Each keeper needs to be mapped to a unique shard.
	// We sort the keeper ids; so same-indexed shards will be sent
	// to their appropriate keeper instances.
	sortedKeys := make([]string, 0, len(keepers))
	for k := range keepers {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	matchingIndex := -1
	for i, key := range sortedKeys {
		if key == id {
			matchingIndex = i
			break
		}
	}

	if matchingIndex == -1 {
		return nil
	}

	if matchingIndex < 0 || matchingIndex >= len(shares) {
		return nil
	}

	return &shares[matchingIndex]
}
