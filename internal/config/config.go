//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package config provides configuration-related functionalities
// for the SPIKE system, including version constants and directory
// management for storing encrypted backups and secrets securely.
package config

import (
	"os"
	"path/filepath"

	"github.com/spiffe/spike/app"
)

var NexusVersion = app.Version
var PilotVersion = app.Version
var KeeperVersion = app.Version
var BootstrapVersion = app.Version

// NexusDataFolder returns the path to the directory where Nexus stores
// its encrypted backup for its secrets and other data.
func NexusDataFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	spikeDir := filepath.Join(homeDir, ".spike")

	// Create the directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create db files in it.
	err = os.MkdirAll(spikeDir+"/data", 0700)
	if err != nil {
		panic(err)
	}

	// The data dir is not configurable for security reasons.
	return filepath.Join(spikeDir, "/data")
}

// PilotRecoveryFolder returns the path to the directory where the
// recovery shards will be stored as a result of the `spike recover`
// command.
func PilotRecoveryFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	spikeDir := filepath.Join(homeDir, ".spike")

	// Create the directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create recovery files in it.
	err = os.MkdirAll(spikeDir+"/recover", 0700)
	if err != nil {
		panic(err)
	}

	// The data dir is not configurable for security reasons.
	return filepath.Join(spikeDir, "/recover")
}
