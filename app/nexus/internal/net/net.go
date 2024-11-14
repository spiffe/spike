//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"errors"

	"github.com/go-jose/go-jose/v4/json"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// UpdateCache sends a cache update request to SPIKE Keep using mutual
// TLS authentication. It sends the current root key value to
// SPIKE Keep's cache endpoint for synchronization.
//
// The function performs the following steps:
// 1. Creates an mTLS client using the provided X509 source
// 2. Constructs a cache request with the current root key
// 3. Sends a POST request to the server's cache endpoint
//
// Parameters:
//   - source: An X509Source containing the necessary certificates for mTLS
//     authentication. -- Must not be nil.
//
// Returns:
//   - error: Returns nil on successful cache update. Otherwise,
//     returns an error if:
//   - The provided source is nil
//   - Failed to create mTLS client
//   - Failed to marshal the request
//   - The POST request failed
//
// Example usage:
//
//	err := UpdateCache(x509Source)
//	if err != nil {
//	    log.Printf("Failed to update cache: %v", err)
//	}
func UpdateCache(source *workloadapi.X509Source, rootKeyFromState string) error {
	if source == nil {
		return errors.New("UpdateCache: got nil source")
	}

	client, err := net.CreateMtlsClient(source, auth.CanTalkToNexus)
	if err != nil {
		return err
	}

	rr := reqres.RootKeyCacheRequest{
		RootKey: rootKeyFromState,
	}

	md, err := json.Marshal(rr)
	if err != nil {
		return errors.New(
			"UpdateCache: failed to marshal request: " + err.Error(),
		)
	}

	_, err = net.Post(client, UrlKeepWrite(), md)
	return err
}

// FetchFromCache retrieves a root key from SPIKE Keeper using an X509 source
// for authentication. It creates an mTLS client using the provided source,
// sends a read request to SPIKE Keeper and returns the root key from the
// response.
//
// Parameters:
//   - source: A pointer to workloadapi.X509Source used for creating the mTLS client.
//     Must not be nil.
//
// Returns:
//   - string: The root key retrieved from the cache.
//   - error: An error if any step fails:
//   - If source is nil
//   - If mTLS client creation fails
//   - If request marshaling fails
//   - If HTTP POST request fails
//   - If response unmarshaling fails
//
// Example:
//
//	source := &workloadapi.X509Source{...}
//	rootKey, err := FetchFromCache(source)
//	if err != nil {
//	    log.Fatalf("failed to fetch root key: %v", err)
//	}
func FetchFromCache(source *workloadapi.X509Source) (string, error) {
	if source == nil {
		return "", errors.New("FetchFromCache: got nil source")
	}

	client, err := net.CreateMtlsClient(source, auth.CanTalkToNexus)
	// client, err := net.CreateMtlsClient(source, config.CanTalkToAnyone)
	if err != nil {
		return "", err
	}

	rr := reqres.RootKeyReadRequest{}

	md, err := json.Marshal(rr)
	if err != nil {
		return "", errors.New(
			"FetchFromCache: failed to marshal request: " + err.Error(),
		)
	}

	data, err := net.Post(client, UrlKeepRead(), md)
	var res reqres.RootKeyReadResponse

	if len(data) == 0 {
		return "", nil
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", errors.New(
			"FetchFromCache: failed to unmarshal response: " + err.Error(),
		)
	}

	return res.RootKey, err
}
