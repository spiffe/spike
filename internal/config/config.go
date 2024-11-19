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

// SpikePilotAdminTokenFile returns the file path where the SPIKE Pilot admin
// JWT should be stored. The function creates the necessary directory structure
// if it doesn't exist.
//
// The function attempts to create a .spike directory in the user's home
// directory. If the home directory cannot be determined, it falls back to
// using /tmp.
//
// The directory is created with 0600 permissions for security.
//
// The token file path and name are hardcoded for security reasons and cannot be
// configured by the user.
//
// Returns the absolute path to the admin JWT token file.
//
// The function will panic if it fails to create the required directory
// structure.
func SpikePilotAdminTokenFile() string {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	// Create path for .spike folder
	spikeDir := filepath.Join(homeDir, ".spike")

	// Create directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create JWTs in it.
	err = os.MkdirAll(spikeDir, 0700)
	if err != nil {
		panic(err)
	}

	// The file path and file name are NOT configurable for security reasons.

	return filepath.Join(spikeDir, ".spike-admin.jwt")
}

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
