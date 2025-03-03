//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"encoding/json"
	"os"
)

// Keepers retrieves and parses the keeper peer configurations from environment.
// It reads SPIKE_NEXUS_KEEPER_PEERS environment variable which should contain a
// JSON map of keeper IDs to their URLs.
//
// The environment variable should be formatted as:
//
//	{
//	    "1": "https://localhost:8443",
//	    "2": "https://localhost:8543",
//	    "3": "https://localhost:8643"
//	}
//
// Returns:
//   - map[string]string: Mapping of keeper IDs to their URLs
//
// Panics if:
//   - SPIKE_NEXUS_KEEPER_PEERS is not set
//   - SPIKE_NEXUS_KEEPER_PEERS contains invalid JSON
func Keepers() map[string]string {
	p := os.Getenv("SPIKE_NEXUS_KEEPER_PEERS")

	if p == "" {
		panic("SPIKE_NEXUS_KEEPER_PEERS has to be configured in the environment")
	}

	// Parse the JSON-formatted environment variable
	peers := make(map[string]string)
	err := json.Unmarshal([]byte(p), &peers)
	if err != nil {
		panic("SPIKE_NEXUS_KEEPER_PEERS contains invalid JSON: " + err.Error())
	}

	return peers
}
