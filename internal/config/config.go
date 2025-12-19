//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package config provides configuration-related functionalities
// for the SPIKE system, including version constants and directory
// management for storing encrypted backups and secrets securely.
package config

// NexusDataFolder returns the path to the directory where Nexus stores
// its encrypted backup for its secrets and other data.
//
// The directory can be configured via the SPIKE_NEXUS_DATA_DIR environment
// variable. If not set or invalid, it falls back to ~/.spike/data.
// If the home directory is unavailable, it falls back to
// /tmp/.spike-$USER/data.
//
// The directory is created once on the first call and cached for the following
// calls.
//
// Returns:
//   - string: The absolute path to the Nexus data directory.
func NexusDataFolder() string {
	nexusDataOnce.Do(func() {
		nexusDataPath = initNexusDataFolder()
	})
	return nexusDataPath
}

// PilotRecoveryFolder returns the path to the directory where the
// recovery shards will be stored as a result of the `spike recover`
// command.
//
// The directory can be configured via the SPIKE_PILOT_RECOVERY_DIR
// environment variable. If not set or invalid, it falls back to
// ~/.spike/recover. If the home directory is unavailable, it falls back to
// /tmp/.spike-$USER/recover.
//
// The directory is created once on the first call and cached for the following
// calls.
//
// Returns:
//   - string: The absolute path to the Pilot recovery directory.
func PilotRecoveryFolder() string {
	pilotRecoveryOnce.Do(func() {
		pilotRecoveryPath = initPilotRecoveryFolder()
	})
	return pilotRecoveryPath
}
