//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"encoding/json"
	"os"
)

func KeeperId() string {
	p := os.Getenv("SPIKE_KEEPER_ID")

	if p == "" {
		panic("SPIKE_KEEPER_ID has to be configured in the environment")
	}

	return p
}

// KeeperRandomSeed retrieves the "SPIKE_KEEPER_RANDOM_SEED" env variable.
// If the variable is not set, it will panic with an error message.
//
// Make sure that this value is unique and shared by all SPIKE Keeper instances.
// Here is a cryptographically secure way to generate such a seed:
//
//	head -c 32 /dev/urandom | xxd -p
func KeeperRandomSeed() string {
	p := os.Getenv("SPIKE_KEEPER_RANDOM_SEED")

	if p == "" {
		panic("SPIKE_KEEPER_RANDOM_SEED has to be configured in the environment")
	}

	return p
}

func Peers() map[string]string {
	// example:
	// '"{"1":"https://localhost:8443",
	//    "2":"https://localhost:8543"
	//    "3":"https://localhost:8643"}

	p := os.Getenv("SPIKE_KEEPER_PEERS")

	if p == "" {
		panic("SPIKE_KEEPER_PEERS has to be configured in the environment")
	}

	// Parse the JSON-formatted environment variable
	peers := make(map[string]string)
	err := json.Unmarshal([]byte(p), &peers)
	if err != nil {
		panic("SPIKE_KEEPER_PEERS contains invalid JSON: " + err.Error())
	}

	return peers
}

func StateFileName() string {
	return "keeper-" + KeeperId() + ".state"
}
