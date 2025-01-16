//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"os"
	"strings"

	"github.com/spiffe/spike/app/keeper/internal/env"
)

type AppState string

// TODO: these states have changed and maybe not even used anymore,
// update the code accordingly.
const AppStateNotReady AppState = "NOT_READY"
const AppStateReady AppState = "READY"
const AppStateRecovering AppState = "RECOVERING"
const AppStateError AppState = "ERROR"

// ReadAppState retrieves and parses the application state from the state file.
// The state file path is determined by env.StateFileName().
//
// The function will return:
//   - AppStateNotReady if the state file doesn't exist or is empty
//   - AppStateError if there was an error reading the file
//   - The trimmed content of the file cast to AppState otherwise
//
// The state data is expected to be a single string value that can be cast
// directly to the AppState type after whitespace trimming.
//
// Returns:
//   - AppState: The current application state
func ReadAppState() AppState {
	data, err := os.ReadFile(env.StateFileName())
	if os.IsNotExist(err) {
		return AppStateNotReady
	}
	if err != nil {
		return AppStateError
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return AppStateNotReady
	}
	return AppState(strings.TrimSpace(string(data)))
}
