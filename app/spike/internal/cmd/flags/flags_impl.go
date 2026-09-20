//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\ SPDX-License-Identifier: Apache-2.0

package flags

import (
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// readFailure builds the error reported when a flag cannot be read.
//
// Parameters:
//   - name: The flag name without the leading dashes
//   - err: The error Cobra returned while reading the flag
//
// Returns:
//   - *sdkErrors.SDKError: ErrDataInvalidInput wrapping err, with a message
//     that names the flag
func readFailure(name string, err error) *sdkErrors.SDKError {
	failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
	failErr.Msg = "failed to read the --" + name + " flag"
	return failErr
}
