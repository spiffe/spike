//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

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
func IsPilot(spiffeid string) bool {
	return spiffeid == SpikePilotSpiffeId()
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
func IsKeeper(spiffeid string) bool {
	return spiffeid == SpikeKeeperSpiffeId()
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
func IsNexus(spiffeid string) bool {
	return spiffeid == SpikeNexusSpiffeId()
}

// CanTalkToAnyone is used for debugging purposes
func CanTalkToAnyone(_ string) bool {
	return true
}

// CanTalkToNexus checks if the provided SPIFFE ID matches the SPIKE Keeper
// SPIFFE ID or SPIKE Pilot SPIFFE ID.
//
// This is used as a validator in SPIKE Nexus, because currently only SPIKE
// Keeper or SPIKE Pilot can communicate with SPIKE Nexus.
//
// This trust will evolve later.
//
// The function compares the input SPIFFE ID against the configured Spike Keeper
// SPIFFE ID from the environment. This is typically used for validating whether
// a given identity represents the Keeper service.
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the SPIFFE ID matches the Keeper SPIFFE ID, false otherwise
func CanTalkToNexus(spiffeid string) bool {
	return spiffeid == SpikeKeeperSpiffeId() || spiffeid == SpikePilotSpiffeId()
}

// CanTalkToKeeper checks if the provided SPIFFE ID matches the SPIKE Nexus
// SPIFFE ID.
//
// This is used as a validator in SPIKE Keeper, because currently only SPIKE
// Nexus can talk to SPIKE Keeper.
//
// The function compares the input SPIFFE ID against the configured Spike Keeper
// SPIFFE ID from the environment. This is typically used for validating whether
// a given identity represents the Keeper service.
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the SPIFFE ID matches the Keeper SPIFFE ID, false otherwise
func CanTalkToKeeper(spiffeid string) bool {
	return spiffeid == SpikeNexusSpiffeId()
}

// CanTalkToPilot checks if the provided SPIFFE ID matches the SPIKE Nexus
// SPIFFE ID.
//
// This is used as a validator in SPIKE Pilot, because currently only SPIKE
// Nexus can talk to SPIKE Pilot.
//
// The function compares the input SPIFFE ID against the configured Spike Pilot
// SPIFFE ID from the environment. This is typically used for validating whether
// a given identity represents the Keeper service.
//
// Parameters:
//   - spiffeid: The SPIFFE ID string to check
//
// Returns:
//   - bool: true if the SPIFFE ID matches the Nexus SPIFFE ID, false otherwise
func CanTalkToPilot(spiffeid string) bool {
	return spiffeid == SpikeNexusSpiffeId()
}
