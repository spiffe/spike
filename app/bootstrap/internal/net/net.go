//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/internal/net"
)

// PutShardContributionRequest sends an HTTP POST request to the specified
// URL using the provided client and payload data. The function is designed for
// sending shard contribution requests to keepers in a secure manner. It
// will terminate the program with exit code 1 if the POST request fails.
func PutShardContributionRequest(
	client *http.Client, u string, md []byte, keeperID string,
) error {
	const fName = "PutShardContributionRequest"

	log.Log().Info(fName, "payload_sha", fmt.Sprintf("%x", sha256.Sum256(md)))

	_, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Info(
			fName,
			"message", "failed to post",
			"err", err,
			"keeper_id", keeperID,
		)
		return err
	}

	return nil
}
