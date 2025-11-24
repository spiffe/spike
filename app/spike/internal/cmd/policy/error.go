//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// handleAPIError processes API errors and prints appropriate messages.
// It helps standardize error handling across policy commands.
//
// Parameters:
//   - cmd: Cobra command for output
//   - err: The error returned from an API call
//
// Returns:
//   - bool: true if an error was handled, false if no error
//
// Usage example:
//
//	policies, err := api.ListPolicies()
//	if handleAPIError(cmd, err) {
//	    return
//	}
func handleAPIError(cmd *cobra.Command, err error) bool {
	if err == nil {
		return false
	}

	if errors.NotReadyError(err) {
		stdout.PrintNotReady()
		return true
	}

	if strings.Contains(err.Error(), "unexpected end of JSON") ||
		strings.Contains(err.Error(), "parsing") {
		cmd.PrintErrln("Error: Failed to parse API response. " +
			"The server may be unavailable or returned an invalid response.")
		cmd.PrintErrf("Technical details: %v\n", err)
		return true
	}

	cmd.PrintErrf("Error: %v\n", err)
	return true
}
