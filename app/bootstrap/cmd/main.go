package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike/app/bootstrap/internal/env"
)

// sanityCheck verifies that a set of secret shares can correctly reconstruct
// the original secret. It performs this verification by attempting to recover
// the secret using the minimum required number of shares and comparing the
// result with the original secret.
//
// Parameters:
//   - secret group.Scalar: The original secret to verify against
//   - shares []shamir.Share: The generated secret shares to verify
//
// The function will:
//   - Calculate the threshold (t) from the environment configuration
//   - Attempt to reconstruct the secret using exactly t+1 shares
//   - Compare the reconstructed secret with the original
//   - Zero out the reconstructed secret regardless of success or failure
//
// If the verification fails, the function will:
//   - Log a fatal error and exit if recovery fails
//   - Log a fatal error and exit if the recovered secret doesn't match the
//     original
//
// Security:
//   - The reconstructed secret is always zeroed out to prevent memory leaks
//   - In case of fatal errors, the reconstructed secret is explicitly zeroed
//     before logging since deferred functions won't run after log.FatalLn
func sanityCheck(secret group.Scalar, shares []shamir.Share) {
	const fName = "sanityCheck"

	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct

	reconstructed, err := shamir.Recover(t, shares[:env.ShamirThreshold()])
	// Security: Ensure that the secret is zeroed out if the check fails.
	defer func() {
		if reconstructed == nil {
			return
		}
		reconstructed.SetUint64(0)
	}()

	if err != nil {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(fName + ": Recovered secret does not match original")
	}
}

func main() {
	const fName = "bootstrap.main"

	var rootKeySeed [32]byte
	// Security: Ensure the rootKeySeed is zeroed out after use.
	defer func() {
		mem.ClearRawBytes(&rootKeySeed)
	}()

	if _, err := rand.Read(rootKeySeed[:]); err != nil {
		log.Fatal(err.Error())
	}

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirShares())        // Total number of shares

	log.Log().Info(fName, "t", t, "n", n)

	// Create a secret from our 32-byte key:
	rootSecret := g.NewScalar()

	if err := rootSecret.UnmarshalBinary(rootKeySeed[:]); err != nil {
		log.FatalLn(fName + ": Failed to unmarshal key: %v" + err.Error())
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rootKeySeed[:])
	ss := shamir.New(reader, t, rootSecret)

	log.Log().Info(fName, "message", "Generated Shamir shares")

	rootShares := ss.Share(n)

	// Security: Ensure the root key and shares are zeroed out after use.
	sanityCheck(rootSecret, rootShares)
	defer func() {
		rootSecret.SetUint64(0)
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}
	}()

	var result = make(map[int]*[32]byte)

	for _, share := range rootShares {
		log.Log().Info(fName, "message", "Generating share", "share.id", share.ID)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to marshal share")
			return
		}

		if len(contribution) != 32 {
			log.Log().Error(fName, "message", "Length of share is unexpected")
			return
		}

		bb, err := share.ID.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to unmarshal share ID")
			return
		}

		bigInt := new(big.Int).SetBytes(bb)
		ii := bigInt.Uint64()

		if len(contribution) != 32 {
			log.Log().Error(fName, "message", "Length of share is unexpected")
			return
		}

		var rs [32]byte
		copy(rs[:], contribution)

		log.Log().Info(fName, "message", "Generated shares", "len", len(rs))

		result[int(ii)] = &rs
	}

	log.Log().Info(fName,
		"message", "Successfully generated pilot recovery shards.")

	fmt.Println("result", result)
}
