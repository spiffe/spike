package env

import (
	"encoding/json"
	"os"
)

func Keepers() map[string]string {
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
