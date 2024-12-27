//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"encoding/json"
	"github.com/spiffe/spike/internal/log"
	"os"
)

// KeeperId retrieves the keeper's unique identifier from the SPIKE_KEEPER_ID
// environment variable. This ID is used to identify this keeper instance within
// the peer network and in state file naming.
//
// The function will panic if SPIKE_KEEPER_ID is not set in the environment, as
// this is a required configuration for keeper operation.
//
// Returns:
//   - string: The keeper ID value from the environment
//
// Panics:
//   - If SPIKE_KEEPER_ID environment variable is not set
func KeeperId() string {
	p := os.Getenv("SPIKE_KEEPER_ID")

	if p == "" {
		panic("SPIKE_KEEPER_ID has to be configured in the environment")
	}

	return p
}

// Peers returns a mapping of keeper IDs to their network URLs, retrieved from
// the SPIKE_KEEPER_PEERS environment variable. This configuration defines the
// peer network topology that this keeper instance will communicate with.
//
// The SPIKE_KEEPER_PEERS value must be a JSON string containing a map where:
//   - Keys are keeper ID strings
//   - Values are the corresponding keeper's full URL including protocol and port
//
// Example JSON format:
//
//	{
//	  "1": "https://localhost:8443",
//	  "2": "https://localhost:8543",
//	  "3": "https://localhost:8643"
//	}
//
// Returns:
//   - map[string]string: Mapping of keeper IDs to their URLs
//
// Panics:
//   - If SPIKE_KEEPER_PEERS environment variable is not set
//   - If the environment variable contains invalid JSON
//   - If JSON unmarshaling fails for any other reason
func Peers() map[string]string {
	// example:
	// '"{"1":"https://localhost:8443",
	//    "2":"https://localhost:8543"
	//    "3":"https://localhost:8643"}

	p := os.Getenv("SPIKE_KEEPER_PEERS")

	if p == "" {
		log.FatalLn("SPIKE_KEEPER_PEERS has to be configured in the environment")
	}

	// Parse the JSON-formatted environment variable
	peers := make(map[string]string)
	err := json.Unmarshal([]byte(p), &peers)
	if err != nil {
		log.FatalLn("SPIKE_KEEPER_PEERS contains invalid JSON: " + err.Error())
	}

	return peers
}

// StateFileName generates the filename used for persisting this keeper's state
// to disk. The filename is constructed by combining the prefix "keeper-" with
// the keeper's ID (from KeeperId()) and the suffix ".state"
//
// For example, if KeeperId() returns "1", the resulting filename would be:
// "keeper-1.state"
//
// Returns:
//   - string: The constructed state filename for this keeper instance
//
// Note: This function depends on KeeperId() and will panic under the same
// conditions as KeeperId().
func StateFileName() string {
	return "keeper-" + KeeperId() + ".state"
}
