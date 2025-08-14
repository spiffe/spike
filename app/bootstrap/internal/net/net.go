//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/internal/net"
)

// Source creates and returns a new SPIFFE X509Source for workload API
// communication. It establishes a connection to the SPIFFE workload API using
// the default endpoint socket. The function will terminate the program with
// exit code 1 if the source creation fails.
func Source() *workloadapi.X509Source {
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

// MTLSClient creates an HTTP client configured for mutual TLS authentication
// using the provided X509Source. The client is configured with a predicate that
// validates peer IDs against the trusted keeper root. Only peers that pass the
// spiffeid.IsKeeper validation will be accepted for connections. The function
// will terminate the program with exit code 1 if client creation fails.
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

// Payload marshals a secret sharing contribution into a JSON payload for
// transmission to a Keeper. It takes a secret sharing share and the target
// Keeper ID, validates the contribution is exactly 32 bytes, and returns the
// marshaled ShardContributionRequest as a byte slice. The function will
// terminate the program with exit code 1 if marshaling fails or if the
// contribution length is invalid.
func Payload(share secretsharing.Share, keeperID string) []byte {
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

// Post sends an HTTP POST request to the specified URL using the provided
// client and payload data. The function is designed for sending shard
// contribution requests to keepers in a secure manner. It will terminate the
// program with exit code 1 if the POST request fails.
func Post(client *http.Client, u string, md []byte, keeperID string) {
	const fName = "post"

	fmt.Printf("POST: %x\n", sha256.Sum256(md))

	_, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn(fName, "message",
			"Failed to post",
			"err", err, "keeper_id", keeperID)
		os.Exit(1)
	}

	fmt.Println("POST: done")
}
