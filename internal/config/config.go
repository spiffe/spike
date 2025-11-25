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

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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
//
// Parameters:
//   - dir: The directory path to validate.
//
// Returns:
//   - *sdkErrors.SDKError: An error if the directory is invalid, restricted,
//     or cannot be accessed. Returns nil if the directory is valid.
func validateDataDirectory(dir string) *sdkErrors.SDKError {
	fName := "validateDataDirectory"

	if dir == "" {
		failErr := *sdkErrors.ErrFSInvalidDirectory // copy
		failErr.Msg = "directory path cannot be empty"
		return &failErr
	}

	// Resolve to an absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		failErr := *sdkErrors.ErrFSInvalidDirectory // copy
		failErr.Msg = fmt.Sprintf("failed to resolve directory path: %s", err)
		return &failErr
	}

	// Check for restricted paths
	for _, restricted := range restrictedPaths {
		if absPath == restricted || strings.HasPrefix(absPath, restricted+"/") {
			failErr := *sdkErrors.ErrFSInvalidDirectory // copy
			failErr.Msg = "path is restricted for security reasons"
			return &failErr
		}
	}

	// Check if using /tmp without user isolation
	if strings.HasPrefix(absPath, "/tmp/") && !strings.Contains(
		absPath, os.Getenv("USER"),
	) {
		log.Warn(fName,
			"message", "Using /tmp without user isolation is not recommended",
			"path", absPath,
		)
	}

	// Check if the directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// copy; TODO: new error type: ErrFSDirectoryDoesNotExist
			failErr := *sdkErrors.ErrFSInvalidDirectory
			failErr.Msg = fmt.Sprintf("failed to check directory: %s", err)
			return &failErr
		}
		// Directory doesn't exist, check if the parent exists, and we can create it
		parentDir := filepath.Dir(absPath)
		if _, err := os.Stat(parentDir); err != nil {
			failErr := *sdkErrors.ErrFSParentDirectoryDoesNotExist // copy
			failErr.Msg = fmt.Sprintf("parent directory does not exist: %s", err)
			return &failErr
		}
	} else {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			failErr := *sdkErrors.ErrFSFileIsNotADirectory // copy
			failErr.Msg = fmt.Sprintf("path is not a directory: %s", absPath)
			return &failErr
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
//
// Returns:
//   - string: The absolute path to the Nexus data directory.
func NexusDataFolder() string {
	const fName = "NexusDataFolder"

	// Check for custom data directory from the environment
	if customDir := os.Getenv("SPIKE_NEXUS_DATA_DIR"); customDir != "" {
		if err := validateDataDirectory(customDir); err == nil {
			// Ensure the directory exists with proper permissions
			dataPath := filepath.Join(customDir, "data")
			if err := os.MkdirAll(dataPath, 0700); err != nil {
				failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
				failErr.Msg = fmt.Sprintf(
					"failed to create custom data directory: %s", err,
				)
				log.WarnErr(fName, *failErr)
			} else {
				return dataPath
			}
		} else {
			failErr := sdkErrors.ErrFSInvalidDirectory.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"invalid custom data directory: %s. using default", customDir,
			)
			log.WarnErr(fName, *failErr)
		}
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to temp with user isolation
		user := os.Getenv("USER")

		if user == "" {
			user = "spike" // TODO: constant.
		}

		tempDir := fmt.Sprintf("/tmp/.spike-%s", user)
		dataPath := filepath.Join(tempDir, "data")
		err = os.MkdirAll(dataPath, 0700)

		if err != nil {
			failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to create temp data directory: %s", err,
			)
			log.FatalErr(fName, *failErr)
		}

		return dataPath
	}

	spikeDir := filepath.Join(homeDir, ".spike") // TODO: constant.
	dataPath := filepath.Join(spikeDir, "data")  // TODO: constant.

	// Create the directory if it doesn't exist
	// 0700 because we want to restrict access to the directory
	// but allow the user to create db files in it.
	err = os.MkdirAll(dataPath, 0700)
	if err != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to create spike data directory: %s", err,
		)
		log.FatalErr(fName, *failErr)
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
//
// Returns:
//   - string: The absolute path to the Pilot recovery directory.
func PilotRecoveryFolder() string {
	const fName = "PilotRecoveryFolder"

	// Check for custom recovery directory from environment
	if customDir := os.Getenv(env.PilotRecoveryDir); customDir != "" { // TODO: don't we have an env package that has helpers for these?
		if err := validateDataDirectory(customDir); err == nil {
			// Ensure the directory exists with proper permissions
			recoverPath := filepath.Join(customDir, "recover")
			if err := os.MkdirAll(recoverPath, 0700); err != nil {
				warnErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
				warnErr.Msg = "failed to create custom recovery directory"
				log.WarnErr(fName, *warnErr)
			} else {
				return recoverPath
			}
		} else {
			warnErr := sdkErrors.ErrFSInvalidDirectory.Wrap(err)
			warnErr.Msg = "invalid custom recovery directory"
			log.WarnErr(fName, *warnErr)
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
			failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
			failErr.Msg = "failed to create temp recovery directory"
			log.FatalErr(fName, *failErr)
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
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
		failErr.Msg = "failed to create spike recovery directory"
		log.FatalErr(fName, *failErr)
	}

	return recoverPath
}

// ValidPermissions contains the set of valid policy permissions supported by
// the SPIKE system. These are sourced from the SDK to prevent typos.
//
// Valid permissions are:
//   - read: Read access to resources
//   - write: Write access to resources
//   - list: List access to resources
//   - execute: Execute access to resources
//   - super: Superuser access (grants all permissions)
var ValidPermissions = []data.PolicyPermission{
	data.PermissionRead,
	data.PermissionWrite,
	data.PermissionList,
	data.PermissionExecute,
	data.PermissionSuper,
}

// validPermission checks if the given permission string is valid.
//
// Parameters:
//   - perm: The permission string to validate.
//
// Returns:
//   - true if the permission is found in ValidPermissions, false otherwise.
func validPermission(perm string) bool {
	for _, p := range ValidPermissions {
		if string(p) == perm {
			return true
		}
	}
	return false
}

// validPermissionsList returns a comma-separated string of valid permissions,
// suitable for display in error messages.
//
// Returns:
//   - string: A comma-separated list of valid permissions.
func validPermissionsList() string {
	perms := make([]string, len(ValidPermissions))
	for i, p := range ValidPermissions {
		perms[i] = string(p)
	}
	return strings.Join(perms, ", ")
}

// ValidatePermissions validates policy permissions from a comma-separated
// string and returns a slice of PolicyPermission values. It returns an error
// if any permission is invalid or if the string contains no valid permissions.
//
// Valid permissions are:
//   - read: Read access to resources
//   - write: Write access to resources
//   - list: List access to resources
//   - execute: Execute access to resources
//   - super: Superuser access (grants all permissions)
//
// Parameters:
//   - permsStr: Comma-separated string of permissions
//     (e.g., "read,write,execute")
//
// Returns:
//   - []data.PolicyPermission: Validated policy permissions
//   - *sdkErrors.SDKError: An error if any permission is invalid
func ValidatePermissions(permsStr string) (
	[]data.PolicyPermission, *sdkErrors.SDKError,
) {
	var permissions []string
	for _, p := range strings.Split(permsStr, ",") {
		perm := strings.TrimSpace(p)
		if perm != "" {
			permissions = append(permissions, perm)
		}
	}

	perms := make([]data.PolicyPermission, 0, len(permissions))
	for _, perm := range permissions {
		if !validPermission(perm) {
			failErr := *sdkErrors.ErrAccessInvalidPermission // copy
			failErr.Msg = fmt.Sprintf(
				"invalid permission: '%s'. valid permissions: '%s'",
				perm, validPermissionsList(),
			)
			return nil, &failErr
		}
		perms = append(perms, data.PolicyPermission(perm))
	}

	if len(perms) == 0 {
		failErr := *sdkErrors.ErrAccessInvalidPermission // copy
		failErr.Msg = "no valid permissions specified" +
			". valid permissions are: " + validPermissionsList()
		return nil, &failErr
	}

	return perms, nil
}
