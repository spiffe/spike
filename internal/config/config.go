//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package config provides configuration-related functionalities
// for the SPIKE system, including version constants and directory
// management for storing encrypted backups and secrets securely.
package config

import (
	"fmt"
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// NexusDataFolder returns the path to the directory where Nexus stores
// its encrypted backup for its secrets and other data.
//
// The directory can be configured via the SPIKE_NEXUS_DATA_DIR environment
// variable. If not set or invalid, it falls back to ~/.spike/data.
// If the home directory is unavailable, it falls back to
// /tmp/.spike-$USER/data.
//
// The directory is created once on the first call and cached for following
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
// The directory is created once on first call and cached for subsequent calls.
//
// Returns:
//   - string: The absolute path to the Pilot recovery directory.
func PilotRecoveryFolder() string {
	pilotRecoveryOnce.Do(func() {
		pilotRecoveryPath = initPilotRecoveryFolder()
	})
	return pilotRecoveryPath
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
			failErr := *sdkErrors.ErrAccessInvalidPermission.Clone()
			failErr.Msg = fmt.Sprintf(
				"invalid permission: '%s'. valid permissions: '%s'",
				perm, validPermissionsList(),
			)
			return nil, &failErr
		}
		perms = append(perms, data.PolicyPermission(perm))
	}

	if len(perms) == 0 {
		failErr := *sdkErrors.ErrAccessInvalidPermission.Clone()
		failErr.Msg = "no valid permissions specified" +
			". valid permissions are: " + validPermissionsList()
		return nil, &failErr
	}

	return perms, nil
}
