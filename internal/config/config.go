//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package config provides configuration-related functionalities
// for the SPIKE system, including version constants and directory
// management for storing encrypted backups and secrets securely.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app"
)

var NexusVersion = app.Version
var PilotVersion = app.Version
var KeeperVersion = app.Version
var BootstrapVersion = app.Version

// restrictedPaths contains system directories that should not be used
// for data storage for security and operational reasons.
var restrictedPaths = []string{
	"/", "/etc", "/sys", "/proc", "/dev", "/bin", "/sbin",
	"/usr", "/lib", "/lib64", "/boot", "/root",
}

// validateDataDirectory checks if a directory path is valid and safe to use
// for storing SPIKE data. It ensures the directory exists or can be created,
// has proper permissions, and is not in a restricted location.
func validateDataDirectory(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Resolve to an absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check for restricted paths
	for _, restricted := range restrictedPaths {
		if absPath == restricted || strings.HasPrefix(absPath, restricted+"/") {
			return fmt.Errorf(
				"path %s is restricted for security reasons", absPath,
			)
		}
	}

	// Check if using /tmp without user isolation
	if strings.HasPrefix(absPath, "/tmp/") && !strings.Contains(absPath, os.Getenv("USER")) {
		log.Log().Warn("validateDataDirectory",
			"message", "Using /tmp without user isolation is not recommended",
			"path", absPath,
		)
	}

	// Check if the directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check directory: %w", err)
		}
		// Directory doesn't exist, check if the parent exists, and we can create it
		parentDir := filepath.Dir(absPath)
		if _, err := os.Stat(parentDir); err != nil {
			return fmt.Errorf(
				"parent directory %s does not exist: %w", parentDir, err,
			)
		}
	} else {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", absPath)
		}
	}

	return nil
}

// NexusDataFolder returns the path to the directory where Nexus stores
// its encrypted backup for its secrets and other data.
//
// The directory can be configured via the SPIKE_NEXUS_DATA_DIR environment
// variable. If not set or invalid, it falls back to ~/.spike/data.
// If the home directory is unavailable, it falls back to
// /tmp/.spike-$USER/data.
func NexusDataFolder() string {
	const fName = "NexusDataFolder"

	// Check for custom data directory from the environment
	if customDir := os.Getenv("SPIKE_NEXUS_DATA_DIR"); customDir != "" {
		if err := validateDataDirectory(customDir); err == nil {
			// Ensure the directory exists with proper permissions
			dataPath := filepath.Join(customDir, "data")
			if err := os.MkdirAll(dataPath, 0700); err != nil {
				log.Log().Warn(fName,
					"message", "Failed to create custom data directory",
					"dir", dataPath,
					"err", err.Error(),
				)
			} else {
				return dataPath
			}
		} else {
			log.Log().Warn(fName,
				"message", "Invalid custom data directory, using default",
				"dir", customDir,
				"err", err.Error(),
			)
		}
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to temp with user isolation
		user := os.Getenv("USER")
		if user == "" {
			user = "spike"
		}
		tempDir := fmt.Sprintf("/tmp/.spike-%s", user)
		dataPath := filepath.Join(tempDir, "data")
		err = os.MkdirAll(dataPath, 0700)
		if err != nil {
			panic(err)
		}
		return dataPath
	}

	spikeDir := filepath.Join(homeDir, ".spike")
	dataPath := filepath.Join(spikeDir, "data")

	// Create the directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create db files in it.
	err = os.MkdirAll(dataPath, 0700)
	if err != nil {
		panic(err)
	}

	return dataPath
}

// PilotRecoveryFolder returns the path to the directory where the
// recovery shards will be stored as a result of the `spike recover`
// command.
//
// The directory can be configured via the SPIKE_PILOT_RECOVERY_DIR
// environment variable. If not set or invalid, it falls back to
// ~/.spike/recover. If the home directory is unavailable, it falls back to
// /tmp/.spike-$USER/recover.
func PilotRecoveryFolder() string {
	const fName = "PilotRecoveryFolder"

	// Check for custom recovery directory from environment
	if customDir := os.Getenv("SPIKE_PILOT_RECOVERY_DIR"); customDir != "" {
		if err := validateDataDirectory(customDir); err == nil {
			// Ensure the directory exists with proper permissions
			recoverPath := filepath.Join(customDir, "recover")
			if err := os.MkdirAll(recoverPath, 0700); err != nil {
				log.Log().Warn(fName,
					"message", "Failed to create custom recovery directory",
					"dir", recoverPath,
					"err", err.Error(),
				)
			} else {
				return recoverPath
			}
		} else {
			log.Log().Warn(fName,
				"message", "Invalid custom recovery directory, using default",
				"dir", customDir,
				"err", err.Error(),
			)
		}
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to temp with user isolation
		user := os.Getenv("USER")
		if user == "" {
			user = "spike"
		}
		tempDir := fmt.Sprintf("/tmp/.spike-%s", user)
		recoverPath := filepath.Join(tempDir, "recover")
		err = os.MkdirAll(recoverPath, 0700)
		if err != nil {
			panic(err)
		}
		return recoverPath
	}

	spikeDir := filepath.Join(homeDir, ".spike")
	recoverPath := filepath.Join(spikeDir, "recover")

	// Create the directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create recovery files in it.
	err = os.MkdirAll(recoverPath, 0700)
	if err != nil {
		panic(err)
	}

	return recoverPath
}
