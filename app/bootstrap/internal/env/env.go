//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spiffe/spike-sdk-go/config/env"
)

// ShamirShares returns the total number of shares to be used in Shamir's
// Secret Sharing. It reads the value from the SPIKE_NEXUS_SHAMIR_SHARES
// environment variable.
//
// Returns:
//   - The number of shares specified in the environment variable if it's a
//     valid positive integer
//   - The default value of 3 if the environment variable is unset, empty,
//     or invalid
//
// This determines the total number of shares that will be created when
//
//	splitting a secret.
func ShamirShares() int {
	p := os.Getenv(env.NexusShamirShares)
	if p != "" {
		mv, err := strconv.Atoi(p)
		if err == nil && mv > 0 {
			return mv
		}
	}

	return 3
}

// ShamirThreshold returns the minimum number of shares required to reconstruct
// the secret in Shamir's Secret Sharing scheme.
// It reads the value from the SPIKE_NEXUS_SHAMIR_THRESHOLD environment
// variable.
//
// Returns:
//   - The threshold specified in the environment variable if it's a valid
//     positive integer
//   - The default value of 2 if the environment variable is unset, empty,
//     or invalid
//
// This threshold value determines how many shares are needed to recover the
// original secret. It should be less than or equal to the total number of
// shares (ShamirShares()).
func ShamirThreshold() int {
	p := os.Getenv(env.NexusShamirThreshold)
	if p != "" {
		mv, err := strconv.Atoi(p)
		if err == nil && mv > 0 {
			return mv
		}
	}

	return 2
}

// validURL validates that a URL is properly formatted and uses HTTPS
func validURL(urlStr string) bool {
	pu, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return pu.Scheme == "https" && pu.Host != ""
}

// Keepers retrieves and parses the keeper peer configurations from the
// environment. It reads SPIKE_NEXUS_KEEPER_PEERS environment variable which
// should contain a comma-separated list of keeper URLs.
//
// The environment variable should be formatted as:
// 'https://localhost:8443,https://localhost:8543,https://localhost:8643'
//
// The SPIKE Keeper address mappings will be automatically assigned starting
// with the key "1" and incrementing by 1 for each subsequent SPIKE Keeper.
//
// Returns:
//   - map[string]string: Mapping of keeper IDs to their URLs
//
// Panics if:
//   - SPIKE_NEXUS_KEEPER_PEERS is not set
func Keepers() map[string]string {
	p := os.Getenv(env.NexusKeeperPeers)

	if p == "" {
		panic("SPIKE_NEXUS_KEEPER_PEERS has to be configured in the environment")
	}

	urls := strings.Split(p, ",")

	// Check for duplicate and empty URLs
	urlMap := make(map[string]bool)
	for i, u := range urls {
		trimmedURL := strings.TrimSpace(u)
		if trimmedURL == "" {
			panic(fmt.Sprintf("Keepers: Empty URL found at position %d", i+1))
		}

		// Validate URL format and security
		if !validURL(trimmedURL) {
			panic(
				fmt.Sprintf(
					"Invalid or insecure URL at position %d: %s", i+1,
					trimmedURL),
			)
		}

		if urlMap[trimmedURL] {
			panic("Duplicate keeper URL detected: " + trimmedURL)
		}

		urlMap[trimmedURL] = true
	}

	// The key of the map is the Shamir Shard index (starting from 1), and
	// the value is the Keeper URL that corresponds to that shard index.
	peers := make(map[string]string)
	for i, u := range urls {
		peers[strconv.Itoa(i+1)] = strings.TrimSpace(u)
	}

	return peers
}

// TrustRootForKeeper returns the trust root domain(s) specifically for
// SPIKE Keeper service.
//
// It retrieves the trust root from the SPIKE_TRUST_ROOT_KEEPER environment
// variable. If the environment variable is not set, it returns the default
// value "spike.ist". The return value can be a comma-delimited string of
// multiple trust root domains.
//
// Returns:
//   - A string containing one or more trust root domains for SPIKE Keeper,
//     comma-delimited if multiple
func TrustRootForKeeper() string {
	tr := os.Getenv(env.TrustRootKeeper)
	if tr == "" {
		return "spike.ist"
	}
	return tr
}

// TrustRoot returns the trust root domain(s) for the application.
//
// It retrieves the trust root from the SPIKE_TRUST_ROOT environment variable.
// If the environment variable is not set, it returns the default value
// "spike.ist". The return value can be a comma-delimited string of multiple
// trust root domains.
//
// Returns:
//   - A string containing one or more trust root domains, comma-delimited if
//     multiple
func TrustRoot() string {
	tr := os.Getenv(env.TrustRoot)
	if tr == "" {
		return "spike.ist"
	}
	return tr
}
