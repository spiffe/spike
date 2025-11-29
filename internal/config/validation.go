//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

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
		failErr := sdkErrors.ErrFSInvalidDirectory.Clone()
		failErr.Msg = "directory path cannot be empty"
		return failErr
	}

	// Resolve to an absolute path
	absPath, absErr := filepath.Abs(dir)
	if absErr != nil {
		failErr := sdkErrors.ErrFSInvalidDirectory.Clone()
		failErr.Msg = fmt.Sprintf("failed to resolve directory path: %s", absErr)
		return failErr
	}

	// Check for restricted paths
	for _, restricted := range restrictedPaths {
		if restricted == "/" {
			// Special case: only block the exact root path, not all paths
			if absPath == "/" {
				failErr := sdkErrors.ErrFSInvalidDirectory.Clone()
				failErr.Msg = "path is restricted for security reasons"
				return failErr
			}
			continue
		}
		if absPath == restricted || strings.HasPrefix(absPath, restricted+"/") {
			failErr := sdkErrors.ErrFSInvalidDirectory.Clone()
			failErr.Msg = "path is restricted for security reasons"
			return failErr
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
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			failErr := sdkErrors.ErrFSDirectoryDoesNotExist.Clone()
			failErr.Msg = fmt.Sprintf("failed to check directory: %s", statErr)
			return failErr
		}
		// Directory doesn't exist, check if the parent exists and we can create it
		parentDir := filepath.Dir(absPath)
		if _, parentErr := os.Stat(parentDir); parentErr != nil {
			failErr := sdkErrors.ErrFSParentDirectoryDoesNotExist.Clone()
			failErr.Msg = fmt.Sprintf(
				"parent directory does not exist: %s", parentErr,
			)
			return failErr
		}
	} else {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			failErr := sdkErrors.ErrFSFileIsNotADirectory.Clone()
			failErr.Msg = fmt.Sprintf("path is not a directory: %s", absPath)
			return failErr
		}
	}

	return nil
}
