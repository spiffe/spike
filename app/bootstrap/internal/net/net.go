//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike/internal/net"
)

// PutShardContributionRequest sends an HTTP POST request to the specified
// URL using the provided client and payload data. The function is designed for
// sending shard contribution requests to keepers in a secure manner. It
// will terminate the program with exit code 1 if the POST request fails.
func PutShardContributionRequest(
	client *http.Client, u string, md []byte,
) *sdkErrors.SDKError {
	_, err := net.Post(client, u, md) // TODO: if this is just net.Post; why have a separate function?
	if err != nil {
		return err
	}

	return nil
}
