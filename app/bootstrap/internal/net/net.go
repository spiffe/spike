//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/internal/net"
)

// Payload marshals a secret sharing contribution into a JSON payload for
// transmission to a Keeper. It takes a secret sharing share and the target
// Keeper ID, validates the contribution is exactly 32 bytes, and returns the
// marshaled ShardPutRequest as a byte slice. The function will
// terminate the program with exit code 1 if marshaling fails or if the
// contribution length is invalid.
func Payload(share secretsharing.Share, keeperID string) []byte {
	const fName = "payload"

	contribution, err := share.Value.MarshalBinary()
	if err != nil {
		log.FatalLn(fName, "message", "Failed to marshal share",
			"err", err, "keeper_id", keeperID)
	}

	if len(contribution) != crypto.AES256KeySize {
		log.FatalLn(fName,
			"message", "invalid contribution length",
			"len", len(contribution), "keeper_id", keeperID)
	}

	scr := reqres.ShardPutRequest{}
	shard := new([crypto.AES256KeySize]byte)
	copy(shard[:], contribution)
	scr.Shard = shard

	md, err := json.Marshal(scr)
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to marshal request",
			"err", err, "keeper_id", keeperID)
	}

	return md
}

// PutShardContributionRequest sends an HTTP POST request to the specified URL using the provided
// client and payload data. The function is designed for sending shard
// contribution requests to keepers in a secure manner. It will terminate the
// program with exit code 1 if the POST request fails.
func PutShardContributionRequest(client *http.Client, u string, md []byte, keeperID string) error {
	const fName = "post"

	log.Log().Info(fName, "payload", fmt.Sprintf("%x", sha256.Sum256(md)))

	_, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Info(fName, "message",
			"Failed to post",
			"err", err, "keeper_id", keeperID)
	}
	return err
}

// VerifyPayload creates a JSON payload for the bootstrap verification request.
// It takes a nonce and ciphertext, and returns the marshaled
// BootstrapVerifyRequest as a byte slice. The function will terminate the
// program with exit code 1 if marshaling fails.
func VerifyPayload(nonce, ciphertext []byte) []byte {
	const fName = "verifyPayload"

	request := reqres.BootstrapVerifyRequest{
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}

	md, err := json.Marshal(request)
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to marshal verification request",
			"err", err)
	}

	return md
}

// PostBootstrapVerifyRequest sends an HTTP POST request with verification data to SPIKE Nexus
// and returns the verification response. It logs the request hash for
// debugging purposes. The function returns the response body and any error
// encountered during the request.
func PostBootstrapVerifyRequest(
	client *http.Client, u string, md []byte,
) ([]byte, error) {
	const fName = "postVerify"

	log.Log().Info(fName, "payload", fmt.Sprintf("%x", sha256.Sum256(md)))

	responseBody, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Error(fName, "message",
			"Failed to post verification request",
			"err", err)
		return nil, err
	}

	return responseBody, nil
}
