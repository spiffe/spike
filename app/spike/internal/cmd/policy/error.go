//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// handleAPIError processes API errors and prints appropriate messages.
// It helps standardize error handling across policy commands.
//
// Parameters:
//   - err: The error returned from an API call
//
// Returns:
//   - bool: true if an error was handled, false if no error
//
// Usage example:
//
//	policies, err := api.ListPolicies()
//	if handleAPIError(err) {
//	    return
//	}
func handleAPIError(err error) bool {
	if err == nil {
		return false
	}

	if errors.NotReadyError(err) {
		stdout.PrintNotReady()
		return true
	}

	if strings.Contains(err.Error(), "unexpected end of JSON") ||
		strings.Contains(err.Error(), "parsing") {
		fmt.Println("Error: Failed to parse API response. " +
			"The server may be unavailable or returned an invalid response.")
		fmt.Printf("Technical details: %v\n", err)
		return true
	}

	fmt.Printf("Error: %v\n", err)
	return true
}
