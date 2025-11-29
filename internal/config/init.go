//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// initNexusDataFolder determines and creates the Nexus data directory.
// Called once via sync.Once from NexusDataFolder().
//
// Resolution order:
//  1. SPIKE_NEXUS_DATA_DIR environment variable (if set and valid)
//  2. ~/.spike/data (if the home directory is available)
//  3. /tmp/.spike-$USER/data (fallback)
//
// Returns:
//   - string: The absolute path to the created data directory.
//
// Note: Calls log.FatalErr if directory creation fails in options 2 or 3.
func initNexusDataFolder() string {
	const fName = "initNexusDataFolder"

	// Option 1: Try custom directory from environment variable.
	if path := tryCustomNexusDataDir(fName); path != "" {
		return path
	}

	// Option 2: Try the home directory.
	if path := tryHomeNexusDataDir(fName); path != "" {
		return path
	}

	// Option 3: Fall back to /tmp with user isolation.
	return createTempNexusDataDir(fName)
}

// tryCustomNexusDataDir attempts to use the SPIKE_NEXUS_DATA_DIR environment
// variable to create a custom data directory.
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path, or empty string if the environment
//     variable is not set, invalid, or directory creation fails.
func tryCustomNexusDataDir(fName string) string {
	customDir := os.Getenv(env.NexusDataDir)
	if customDir == "" {
		return ""
	}

	if validateErr := validateDataDirectory(customDir); validateErr != nil {
		failErr := sdkErrors.ErrFSInvalidDirectory.Wrap(validateErr)
		failErr.Msg = fmt.Sprintf(
			"invalid custom data directory: %s. using default", customDir,
		)
		log.WarnErr(fName, *failErr)
		return ""
	}

	dataPath := filepath.Join(customDir, spikeDataFolderName)
	if mkdirErr := os.MkdirAll(dataPath, 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = fmt.Sprintf(
			"failed to create custom data directory: %s", mkdirErr,
		)
		log.WarnErr(fName, *failErr)
		return ""
	}

	return dataPath
}

// tryHomeNexusDataDir attempts to create the data directory under the user's
// home directory (~/.spike/data).
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path, or empty string if the home directory
//     is not available.
//
// Note: Calls log.FatalErr if the home directory exists, but directory creation
// fails.
func tryHomeNexusDataDir(fName string) string {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return ""
	}

	spikeDir := filepath.Join(homeDir, spikeHiddenFolderName)
	dataPath := filepath.Join(spikeDir, spikeDataFolderName)

	// 0700: restrict access to the owner only.
	if mkdirErr := os.MkdirAll(dataPath, 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = fmt.Sprintf(
			"failed to create spike data directory: %s", mkdirErr,
		)
		log.FatalErr(fName, *failErr)
	}

	return dataPath
}

// createTempNexusDataDir creates the data directory under /tmp with user
// isolation. This is the last resort fallback when neither the environment
// variable nor the home directory options are available.
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path (/tmp/.spike-$USER/data).
//
// Note: Calls log.FatalErr if directory creation fails, as this is the final
// fallback option.
func createTempNexusDataDir(fName string) string {
	user := os.Getenv("USER")
	if user == "" {
		user = "spike"
	}

	tempDir := fmt.Sprintf("/tmp/.spike-%s", user)
	dataPath := filepath.Join(tempDir, spikeDataFolderName)

	if mkdirErr := os.MkdirAll(dataPath, 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = fmt.Sprintf(
			"failed to create temp data directory: %s", mkdirErr,
		)
		log.FatalErr(fName, *failErr)
	}

	return dataPath
}

// initPilotRecoveryFolder determines and creates the Pilot recovery directory.
// Called once via sync.Once from PilotRecoveryFolder().
//
// Resolution order:
//  1. SPIKE_PILOT_RECOVERY_DIR environment variable (if set and valid)
//  2. ~/.spike/recover (if the home directory is available)
//  3. /tmp/.spike-$USER/recover (fallback)
//
// Returns:
//   - string: The absolute path to the created recovery directory.
//
// Note: Calls log.FatalErr if directory creation fails in options 2 or 3.
func initPilotRecoveryFolder() string {
	const fName = "initPilotRecoveryFolder"

	// Option 1: Try custom directory from environment variable.
	if path := tryCustomPilotRecoveryDir(fName); path != "" {
		return path
	}

	// Option 2: Try the home directory.
	if path := tryHomePilotRecoveryDir(fName); path != "" {
		return path
	}

	// Option 3: Fall back to /tmp with user isolation.
	return createTempPilotRecoveryDir(fName)
}

// tryCustomPilotRecoveryDir attempts to use the SPIKE_PILOT_RECOVERY_DIR
// environment variable to create a custom recovery directory.
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path, or empty string if the environment
//     variable is not set, invalid, or directory creation fails.
func tryCustomPilotRecoveryDir(fName string) string {
	customDir := os.Getenv(env.PilotRecoveryDir)
	if customDir == "" {
		return ""
	}

	if validateErr := validateDataDirectory(customDir); validateErr != nil {
		warnErr := sdkErrors.ErrFSInvalidDirectory.Wrap(validateErr)
		warnErr.Msg = "invalid custom recovery directory"
		log.WarnErr(fName, *warnErr)
		return ""
	}

	recoverPath := filepath.Join(customDir, spikeRecoveryFolderName)
	if mkdirErr := os.MkdirAll(recoverPath, 0700); mkdirErr != nil {
		warnErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		warnErr.Msg = "failed to create custom recovery directory"
		log.WarnErr(fName, *warnErr)
		return ""
	}

	return recoverPath
}

// tryHomePilotRecoveryDir attempts to create the recovery directory under the
// user's home directory (~/.spike/recover).
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path, or empty string if the home directory
//     is not available.
//
// Note: Calls log.FatalErr if the home directory exists, but directory creation
// fails.
func tryHomePilotRecoveryDir(fName string) string {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return ""
	}

	spikeDir := filepath.Join(homeDir, spikeHiddenFolderName)
	recoverPath := filepath.Join(spikeDir, spikeRecoveryFolderName)

	// 0700: restrict access to the owner only.
	if mkdirErr := os.MkdirAll(recoverPath, 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = "failed to create spike recovery directory"
		log.FatalErr(fName, *failErr)
	}

	return recoverPath
}

// createTempPilotRecoveryDir creates the recovery directory under /tmp with
// user isolation. This is the last resort fallback when neither the environment
// variable nor the home directory options are available.
//
// Parameters:
//   - fName: The caller's function name for logging purposes.
//
// Returns:
//   - string: The created directory path (/tmp/.spike-$USER/recover).
//
// Note: Calls log.FatalErr if directory creation fails, as this is the final
// fallback option.
func createTempPilotRecoveryDir(fName string) string {
	user := os.Getenv("USER")
	if user == "" {
		user = "spike"
	}

	tempDir := fmt.Sprintf("/tmp/.spike-%s", user)
	recoverPath := filepath.Join(tempDir, spikeRecoveryFolderName)

	if mkdirErr := os.MkdirAll(recoverPath, 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = "failed to create temp recovery directory"
		log.FatalErr(fName, *failErr)
	}

	return recoverPath
}
