//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import "os"

func trustRoot() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}

func spikeKeeperSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/keeper"
}

func spikeNexusSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/nexus"
}

func spikePilotSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/pilot/role/superuser"
}
