package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"os"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/app/bootstrap/internal/validation"

	"net/url"
	"strconv"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/internal/net"
)

func keeperShare(rootShares []secretsharing.Share, keeperID string) secretsharing.Share {
	const fName = "keeperShare"

	var share secretsharing.Share
	for _, sr := range rootShares {
		kid, err := strconv.Atoi(keeperID)
		if err != nil {
			log.Log().Warn(
				fName, "message", "Failed to convert keeper id to int", "err", err)
			os.Exit(1)
		}

		if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
			share = sr
			break
		}
	}

	if share.ID.IsZero() {
		log.Log().Warn(fName,
			"message", "Failed to find share for keeper", "keeper_id", keeperID)
		os.Exit(1)
	}

	return share
}

func payload(share secretsharing.Share, keeperID string) []byte {
	const fName = "payload"

	contribution, err := share.Value.MarshalBinary()
	if err != nil {
		log.Log().Warn(fName, "message", "Failed to marshal share",
			"err", err, "keeper_id", keeperID)
		os.Exit(1)
	}

	if len(contribution) != 32 {
		log.Log().Warn(fName,
			"message", "invalid contribution length",
			"len", len(contribution), "keeper_id", keeperID)
		os.Exit(1)
	}

	scr := reqres.ShardContributionRequest{}
	shard := new([32]byte)
	copy(shard[:], contribution)
	scr.Shard = shard

	md, err := json.Marshal(scr)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to marshal request",
			"err", err, "keeper_id", keeperID)
		os.Exit(1)
	}

	return md
}

func keeperEndpoint(keeperAPIRoot string) string {
	const fName = "keeperEndpoint"

	u, err := url.JoinPath(
		keeperAPIRoot, string(apiUrl.KeeperContribute),
	)
	if err != nil {
		log.Log().Warn(
			fName, "message", "Failed to join path", "url", keeperAPIRoot,
		)
		os.Exit(1)
	}
	return u
}

func post(client *http.Client, u string, md []byte, keeperID string) {
	const fName = "post"
	_, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn(fName, "message",
			"Failed to post",
			"err", err, "keeper_id", keeperID)
		os.Exit(1)
	}
}

func MTLSClient(source *workloadapi.X509Source) *http.Client {
	const fName = "MTLSClient"
	client, err := network.CreateMTLSClientWithPredicate(
		source, func(peerId string) bool {
			return spiffeid.IsKeeper(env.TrustRootForKeeper(), peerId)
		},
	)
	if err != nil {
		log.Log().Warn(fName,
			"message", "Failed to create mTLS client",
			"err", err)
		os.Exit(1)
	}
	return client
}

func rootShares() []secretsharing.Share {
	const fName = "rootShares"

	var rootKeySeed [32]byte
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
		os.Exit(1)
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

	rs := ss.Share(n)

	// Security: Ensure the root key and shares are zeroed out after use.
	validation.SanityCheck(rootSecret, rs)

	log.Log().Info(fName, "message", "Successfully generated shards.")
	return rs
}

func source() *workloadapi.X509Source {
	const fName = "source"
	source, _, err := spiffe.Source(
		context.Background(), spiffe.EndpointSocket(),
	)
	if err != nil {
		log.Log().Info(fName, "message", "Failed to create source", "err", err)
		os.Exit(1)
	}
	return source
}

func main() {
	src := source()
	defer spiffe.CloseSource(src)

	for keeperID, keeperAPIRoot := range env.Keepers() {
		post(
			MTLSClient(src),
			keeperEndpoint(keeperAPIRoot),
			payload(keeperShare(rootShares(), keeperID), keeperID),
			keeperID,
		)
	}
}
