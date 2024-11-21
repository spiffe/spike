//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const NexusVersion = "0.1.0"
const PilotVersion = "0.1.0"
const KeeperVersion = "0.1.0"

const NexusIssuer = "spike-nexus"
const NexusAdminSubject = "spike-admin"
const NexusAdminTokenId = "spike-admin-jwt"

func SpikeNexusDataFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	spikeDir := filepath.Join(homeDir, ".spike")

	// Create directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create db files in it.
	err = os.MkdirAll(spikeDir+"/data", 0700)
	if err != nil {
		panic(err)
	}

	// The data dir is not configurable for security reasons.
	return filepath.Join(spikeDir, "/data")
}

func SpikePilotRecoveryFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	spikeDir := filepath.Join(homeDir, ".spike")

	// Create directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create db files in it.
	err = os.MkdirAll(spikeDir+"/recovery", 0700)
	if err != nil {
		panic(err)
	}

	// The data dir is not configurable for security reasons.
	return filepath.Join(spikeDir, "/recovery")
}

func SpikePilotRootKeyRecoveryFile() string {
	folder := SpikePilotRecoveryFolder()

	// The file path and file name are NOT configurable for security reasons.
	return filepath.Join(folder, ".root-key-recovery.spike")
}
