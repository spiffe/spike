//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// validUrl validates that a URL is properly formatted and uses HTTPS
func validUrl(urlStr string) bool {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsedUrl.Scheme == "https" && parsedUrl.Host != ""
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
	p := os.Getenv("SPIKE_NEXUS_KEEPER_PEERS")

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
		if !validUrl(trimmedURL) {
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
