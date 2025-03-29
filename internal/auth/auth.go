//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import "github.com/spiffe/spike-sdk-go/spiffeid"

// TODO: this module can move to spiffe-sdk-go

// IsPilot checks if a given SPIFFE ID matches the SPIKE Pilot's SPIFFE ID.
//
// This function is used for identity verification to determine if the provided
// SPIFFE ID belongs to a SPIKE pilot instance. It compares the input against
// the expected pilot SPIFFE ID returned by SpikePilotSpiffeId().
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the provided SPIFFE ID matches the pilot's ID, false
//     otherwise
//
// Example usage:
//
//	id := "spiffe://example.org/spike/pilot"
//	if IsPilot(id) {
//	    // Handle pilot-specific logic
//	}
func IsPilot(id string) bool {
	return id == spiffeid.SpikePilot()
}

// IsPilotRecover checks if a given SPIFFE ID matches the SPIKE Pilot's
// recovery SPIFFE ID.
//
// This function verifies if the provided SPIFFE ID corresponds to a SPIKE Pilot
// instance with recovery capabilities by comparing it against the expected
// recovery SPIFFE ID returned by SpikePilotRecoverSpiffeId().
//
// Parameters:
//   - spiffeId: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the provided SPIFFE ID matches the pilot's recovery ID,
//     false otherwise
//
// Example usage:
//
//	id := "spiffe://example.org/spike/pilot/recover"
//	if IsPilotRecover(id) {
//	    // Handle recovery-specific logic
//	}
func IsPilotRecover(id string) bool {
	return id == spiffeid.SpikePilotRecover()
}

// IsPilotRestore checks if a given SPIFFE ID matches the SPIKE Pilot's restore
// SPIFFE ID.
//
// This function verifies if the provided SPIFFE ID corresponds to a pilot
// instance with restore capabilities by comparing it against the expected
// restore SPIFFE ID returned by SpikePilotRestoreSpiffeId().
//
// Parameters:
//   - spiffeId: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the provided SPIFFE ID matches the pilot's restore ID,
//     false otherwise
//
// Example usage:
//
//	id := "spiffe://example.org/spike/pilot/restore"
//	if IsPilotRestore(id) {
//	    // Handle restore-specific logic
//	}
func IsPilotRestore(spiffeId string) bool {
	return spiffeId == spiffeid.SpikePilotRestore()
}

// IsKeeper checks if a given SPIFFE ID matches the SPIKE Keeper's SPIFFE ID.
//
// This function is used for identity verification to determine if the provided
// SPIFFE ID belongs to a SPIKE Keeper instance. It compares the input against
// the expected pilot SPIFFE ID returned by SpikeKeeperSpiffeId().
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the provided SPIFFE ID matches the SPIKE Keeper's ID, false
//     otherwise
//
// Example usage:
//
//	id := "spiffe://example.org/spike/keeper"
//	if IsKeeper(id) {
//	    // Handle pilot-specific logic
//	}
func IsKeeper(id string) bool {
	return id == spiffeid.SpikeKeeper()
}

// IsNexus checks if the provided SPIFFE ID matches the SPIKE Nexus SPIFFE ID.
//
// The function compares the input SPIFFE ID against the configured Spike Nexus
// SPIFFE ID from the environment. This is typically used for validating whether
// a given identity represents the Nexus service.
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the SPIFFE ID matches the Nexus SPIFFE ID, false otherwise
func IsNexus(id string) bool {
	return id == spiffeid.SpikeNexus()
}

// PeerCanTalkToAnyone is used for debugging purposes
func PeerCanTalkToAnyone(_ string) bool {
	return true
}

// PeerCanTalkToKeeper checks if the provided SPIFFE ID matches the SPIKE Nexus
// SPIFFE ID.
//
// This is used as a validator in SPIKE Keeper, because currently only SPIKE
// Nexus can talk to SPIKE Keeper.
//
// Parameters:
//   - peerSpiffeId: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the SPIFFE ID matches SPIKE Nexus' SPIFFE ID,
//     false otherwise
func PeerCanTalkToKeeper(peerSpiffeId string) bool {
	return peerSpiffeId == spiffeid.SpikeNexus()
}
