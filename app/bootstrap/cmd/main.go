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
)

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
	secret := g.NewScalar()

	if err := secret.UnmarshalBinary(rootKeySeed[:]); err != nil {
		log.FatalLn(fName + ": Failed to unmarshal key: %v" + err.Error())
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rootKeySeed[:])
	ss := shamir.New(reader, t, secret)

	log.Log().Info(fName, "message", "Generated Shamir shares")

	computedShares := ss.Share(n)

	// secret, computedShares
	// rootSecret, rootShares

	// Security: Ensure the root key and shares are zeroed out after use.
	//sanityCheck(rootSecret, rootShares)
	//defer func() {
	//	rootSecret.SetUint64(0)
	//	for i := range rootShares {
	//		rootShares[i].Value.SetUint64(0)
	//	}
	//}()

	var result = make(map[int]*[32]byte)

	for _, share := range rootShares {
		log.Log().Info(fName, "message", "Generating share", "share.id", share.ID)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to marshal share")
			return nil
		}

		if len(contribution) != 32 {
			log.Log().Error(fName, "message", "Length of share is unexpected")
			return nil
		}

		bb, err := share.ID.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to unmarshal share ID")
			return nil
		}

		bigInt := new(big.Int).SetBytes(bb)
		ii := bigInt.Uint64()

		if len(contribution) != 32 {
			log.Log().Error(fName, "message", "Length of share is unexpected")
			return nil
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
