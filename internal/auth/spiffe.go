//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import "os"

func trustRoot() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT") // TODO: this should be documented and should come from the common env module.
	if tr == "" {
		return "spike.ist"
	}
	return tr
}

func SpikeKeeperSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/keeper"
}

func SpikeNexusSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/nexus"
}

func SpikePilotSpiffeId() string {
	return "spiffe://" + trustRoot() + "/spike/pilot/role/superuser"
}
