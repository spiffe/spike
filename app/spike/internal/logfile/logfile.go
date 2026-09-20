//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package logfile

import (
	"os"
	"path/filepath"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

const (
	// spikeHiddenFolderName is the per-user SPIKE folder, as in the SDK.
	spikeHiddenFolderName = ".spike"
	// fileName is the name of the Pilot diagnostics log.
	fileName = "pilot.log"
	// tempFolderPrefix is the user-isolated fallback under /tmp, as in the
	// SDK's recovery directory resolution.
	tempFolderPrefix = "/tmp/.spike-"
)

// Path returns the well-known location of the SPIKE Pilot diagnostics log.
//
// The path is $HOME/.spike/pilot.log when the home directory can be
// resolved, and /tmp/.spike-$USER/pilot.log otherwise ($USER falls back to
// "spike"). No file is created.
//
// Returns:
//   - string: The absolute path of the diagnostics log file.
func Path() string {
	if homeDir, homeErr := os.UserHomeDir(); homeErr == nil && homeDir != "" {
		return filepath.Join(homeDir, spikeHiddenFolderName, fileName)
	}

	user := os.Getenv("USER")
	if user == "" {
		user = "spike"
	}
	return filepath.Join(tempFolderPrefix+user, fileName)
}

// Route binds the SDK logger to the diagnostics log file.
//
// The SDK logger binds to os.Stdout on first use. Route opens the log file,
// points os.Stdout at it for the duration of that first use, and restores
// os.Stdout afterwards. It must be the first logging-related call in the
// process and must run on the main goroutine before any other goroutine
// starts; a logger that was already initialized keeps its writer.
//
// The file is opened for appending and created with 0600 permissions inside
// a 0700 directory. It stays open for the life of the process because the
// logger owns it.
//
// Returns:
//   - *sdkErrors.SDKError: ErrFSDirectoryCreationFailed when the parent
//     directory cannot be created, ErrFSFileOpenFailed when the file cannot
//     be opened, nil on success.
func Route() *sdkErrors.SDKError {
	path := filepath.Clean(Path())

	// 0700: restrict access to the owner only.
	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0700); mkdirErr != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(mkdirErr)
		failErr.Msg = "failed to create the pilot log directory"
		return failErr
	}

	// 0600: the log may carry error details; owner read/write only.
	file, openErr := os.OpenFile(
		path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600,
	)
	if openErr != nil {
		failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(openErr)
		failErr.Msg = "failed to open the pilot log file"
		return failErr
	}

	stdout := os.Stdout
	os.Stdout = file
	log.Log()
	os.Stdout = stdout

	return nil
}
