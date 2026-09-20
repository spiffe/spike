//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package flags

import (
	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// String returns the value of the named string flag.
//
// Parameters:
//   - cmd: The Cobra command whose flag set holds the flag
//   - name: The flag name without the leading dashes
//
// Returns:
//   - string: The flag value
//   - *sdkErrors.SDKError: ErrDataInvalidInput wrapping the Cobra error if
//     the flag is missing or is not a string flag; nil otherwise
func String(cmd *cobra.Command, name string) (string, *sdkErrors.SDKError) {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return "", readFailure(name, err)
	}
	return value, nil
}

// Int returns the value of the named integer flag.
//
// Parameters:
//   - cmd: The Cobra command whose flag set holds the flag
//   - name: The flag name without the leading dashes
//
// Returns:
//   - int: The flag value
//   - *sdkErrors.SDKError: ErrDataInvalidInput wrapping the Cobra error if
//     the flag is missing or is not an integer flag; nil otherwise
func Int(cmd *cobra.Command, name string) (int, *sdkErrors.SDKError) {
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		return 0, readFailure(name, err)
	}
	return value, nil
}
